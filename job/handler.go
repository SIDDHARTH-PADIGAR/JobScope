package job

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

// POST /jobs
func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Status      string `json:"status"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    int    `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	priority := input.Priority
	if priority < 1 || priority > 5 {
		priority = 3
	}

	job, err := h.Service.CreateJob(input.Title, input.Description, priority)
	if err != nil {
		http.Error(w, "failed to create job", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(job)
}

// GET /jobs
func (h *Handler) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	jobs := h.Service.GetAllJobs()
	json.NewEncoder(w).Encode(jobs)
}

// GET /jobs/{id}
func (h *Handler) GetJobByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}

	job, err := h.Service.GetJobByID(id)
	if err != nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(job)
}

// GET /jobs/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.Service.GetStats()
	json.NewEncoder(w).Encode(stats)
}

// PATCH /jobs/{id}/status
func (h *Handler) UpdateJobStatus(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	job, err := h.Service.UpdateJobStatus(id, input.Status)
	if err != nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(job)
}
