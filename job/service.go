package job

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Stats struct {
	Total   int `json:"total"`
	Queued  int `json:"queued"`
	Running int `json:"running"`
	Done    int `json:"done"`
	Failed  int `json:"failed"`
}

type Service struct {
	jobs     []Job
	mu       sync.Mutex
	nextID   int
	shutdown chan struct{} // for graceful shutdown
}

// NewService initializes the job service and loads from file
func NewService() (*Service, error) {
	jobs, err := LoadJobs()
	if err != nil {
		return nil, err
	}

	// figure out the next ID
	maxID := 0
	for _, job := range jobs {
		if job.ID > maxID {
			maxID = job.ID
		}
	}

	return &Service{
		jobs:     jobs,
		nextID:   maxID + 1,
		shutdown: make(chan struct{}),
	}, nil
}

// CreateJob accepts title, description, priority
func (s *Service) CreateJob(title, desc string, priority int) (*Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := Job{
		ID:          s.nextID,
		Title:       title,
		Description: desc,
		Status:      "queued",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Priority:    priority,
		Retries:     0,
	}

	s.jobs = append(s.jobs, job)
	s.nextID++

	if err := SaveJobs(s.jobs); err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *Service) GetAllJobs() []Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.jobs
}

func (s *Service) GetJobByID(id int) (*Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, job := range s.jobs {
		if job.ID == id {
			j := job
			return &j, nil
		}
	}
	return nil, errors.New("job not found")
}

func (s *Service) UpdateJobStatus(id int, newStatus string) (*Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, job := range s.jobs {
		if job.ID == id {
			s.jobs[i].Status = newStatus
			s.jobs[i].UpdatedAt = time.Now()

			if err := SaveJobs(s.jobs); err != nil {
				return nil, err
			}
			return &s.jobs[i], nil
		}
	}
	return nil, errors.New("job not found")
}

func (s *Service) GetStats() Stats {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats := Stats{}
	for _, job := range s.jobs {
		stats.Total++
		switch job.Status {
		case "queued":
			stats.Queued++
		case "running":
			stats.Running++
		case "done":
			stats.Done++
		case "failed":
			stats.Failed++
		}
	}
	return stats
}

func (s *Service) StartWorker() {
	go func() {
		for {
			select {
			case <-s.shutdown:
				log.Println("[Worker] Shutdown signal received. Exiting worker loop.")
				return
			default:
				s.processNextJob()
				time.Sleep(3 * time.Second)
			}
		}
	}()
}

func (s *Service) StopWorker() {
	close(s.shutdown)
}

func (s *Service) processNextJob() {
	s.mu.Lock()

	var nextIdx = -1
	highestPriority := 6 // valid priority: 1 to 5

	for i, job := range s.jobs {
		if job.Status == "queued" && job.Priority < highestPriority {
			highestPriority = job.Priority
			nextIdx = i
		}
	}

	if nextIdx == -1 {
		s.mu.Unlock()
		return // no queued jobs
	}

	// move to running
	job := &s.jobs[nextIdx]
	job.Status = "running"
	job.UpdatedAt = time.Now()
	fmt.Printf("[Worker] Running job ID %d: %s (retry %d)\n", job.ID, job.Title, job.Retries)

	SaveJobs(s.jobs)
	s.mu.Unlock()

	// simulate job work
	time.Sleep(5 * time.Second)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Simulate failure: 30% chance
	if rand.Intn(100) < 30 {
		job.Retries++
		if job.Retries > 3 {
			job.Status = "failed"
			fmt.Printf("[Worker] Job ID %d failed permanently after 3 retries\n", job.ID)
		} else {
			job.Status = "queued"
			fmt.Printf("[Worker] Job ID %d failed (retry %d), requeued\n", job.ID, job.Retries)
		}
	} else {
		job.Status = "done"
		fmt.Printf("[Worker] Completed job ID %d: %s\n", job.ID, job.Title)
	}

	job.UpdatedAt = time.Now()
	SaveJobs(s.jobs)
}
