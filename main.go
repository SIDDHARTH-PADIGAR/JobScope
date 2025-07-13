package main

import (
	"JobScope/job"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	multi := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multi)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Starting JobScope server...")

	service, err := job.NewService()
	if err != nil {
		log.Fatalf("Failed to load jobs: %v", err)
	}

	handler := job.NewHandler(service)

	// Start background worker
	service.StartWorker()

	// Setup router
	r := mux.NewRouter()
	r.HandleFunc("/jobs", handler.CreateJob).Methods("POST")
	r.HandleFunc("/jobs", handler.GetAllJobs).Methods("GET")
	r.HandleFunc("/jobs/stats", handler.GetStats).Methods("GET") // <- put this before {id}
	r.HandleFunc("/jobs/{id}", handler.GetJobByID).Methods("GET")
	r.HandleFunc("/jobs/{id}/status", handler.UpdateJobStatus).Methods("PATCH")

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
