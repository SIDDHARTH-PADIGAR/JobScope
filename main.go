package main

import (
	"JobScope/job"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// Ensure log directory exists
	os.MkdirAll("logs", os.ModePerm)

	// Setup logging
	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	multi := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multi)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Starting JobScope server...")

	// Initialize service
	service, err := job.NewService()
	if err != nil {
		log.Fatalf("Failed to load jobs: %v", err)
	}
	handler := job.NewHandler(service)

	// Start background worker
	service.StartWorker()

	// Setup routes
	r := mux.NewRouter()

	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if creds.Username != "admin" || creds.Password != "password" {
			http.Error(w, "UNauthorized", http.StatusUnauthorized)
			return
		}

		token, err := job.GenerateToken(creds.Username)
		if err != nil {
			http.Error(w, "Could not generate token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}).Methods("POST")

	r.Handle("/jobs", job.AUthMuddlieware(http.HandlerFunc(handler.CreateJob))).Methods("POST")
	r.HandleFunc("/jobs", handler.GetAllJobs).Methods("GET")
	r.HandleFunc("/jobs/stats", handler.GetStats).Methods("GET") // keep before /{id}
	r.HandleFunc("/jobs/{id}", handler.GetJobByID).Methods("GET")
	r.Handle("/jobs/{id}/status", job.AUthMuddlieware(http.HandlerFunc(handler.UpdateJobStatus))).Methods("PATCH")

	// Setup HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Graceful shutdown listener
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in goroutine
	go func() {
		fmt.Println("Server started at http://localhost:8080")
		log.Println("Listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdownCtx.Done()
	log.Println("Shutdown signal received...")

	// Stop background worker
	service.StopWorker()

	// Shutdown HTTP server gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}
