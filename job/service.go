package job

import (
	"context"
	"errors"
	"log"
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

type Service struct { //acts as a thread-safe queue for jobs between dispatcher and workers
	jobs       []Job
	mu         sync.Mutex
	nextID     int
	shutdown   chan struct{} // for graceful shutdown
	jobQueue   chan Job      // channel queue
	workerPool int           // number of workers
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
		jobs:       jobs,
		nextID:     maxID + 1,
		shutdown:   make(chan struct{}),
		jobQueue:   make(chan Job, 100), // buffered channel
		workerPool: 3,                   // number of concurrent workers,
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

// spawns multiple goroutines via go s.worker(i)
// each worker listens on jobQueue and pulls jobs to excute concurrently.
func (s *Service) StartWorker() {
	for i := 0; i < s.workerPool; i++ {
		go s.worker(i + 1)
	}

	go s.dispatcher()
}

func (s *Service) StopWorker() {
	close(s.shutdown)
}

func (s *Service) processNextJob() {
	s.mu.Lock()
	var selectedIndex = -1
	var highestPriority = 6 // out of range
	for i, job := range s.jobs {
		if job.Status == "queued" && job.Priority < highestPriority {
			selectedIndex = i
			highestPriority = job.Priority
		}
	}
	if selectedIndex == -1 {
		s.mu.Unlock()
		return // no job to process
	}

	// Mark as running
	s.jobs[selectedIndex].Status = "running"
	s.jobs[selectedIndex].UpdatedAt = time.Now()
	jobID := s.jobs[selectedIndex].ID
	title := s.jobs[selectedIndex].Title
	SaveJobs(s.jobs)
	s.mu.Unlock()

	log.Printf("[Worker] Running job ID %d: %s", jobID, title)

	// Simulate work with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan bool)

	go func() {
		time.Sleep(6 * time.Second) // simulate heavy work, force timeout
		done <- true
	}()

	select {
	case <-ctx.Done():
		s.mu.Lock()
		s.jobs[selectedIndex].Status = "failed"
		s.jobs[selectedIndex].UpdatedAt = time.Now()
		SaveJobs(s.jobs)
		s.mu.Unlock()
		log.Printf("[Worker] Job ID %d timed out and failed", jobID)
	case <-done:
		s.mu.Lock()
		s.jobs[selectedIndex].Status = "done"
		s.jobs[selectedIndex].UpdatedAt = time.Now()
		SaveJobs(s.jobs)
		s.mu.Unlock()
		log.Printf("[Worker] Completed job ID %d: %s", jobID, title)
	}
}

// scans s.jobs for queued jobs
// sends them into jobsQueue
// marks them as enqueued to prever duplicates
// runs in a separate goroutine
func (s *Service) dispatcher() {
	for {
		select {
		case <-s.shutdown:
			log.Println("[Dispatcher] Shutdown received")
			return
		default:
			s.mu.Lock()
			for i := range s.jobs {
				if s.jobs[i].Status == "queued" {
					s.jobQueue <- s.jobs[i]
					s.jobs[i].Status = "enqueued"
					s.jobs[i].UpdatedAt = time.Now()
				}
			}
			s.mu.Unlock()
			time.Sleep(2 * time.Second)
		}
	}
}

func (s *Service) worker(id int) { //multiple worker goroutines (via worker pool) pull from jobQueue concurrently
	for { // each worker runs a job inside a timeout-protected context
		select {
		case <-s.shutdown:
			log.Printf("[Worker %d] Shutdown received. Exiting.\n", id)
			return
		case job := <-s.jobQueue:
			s.runJob(id, job)
		}
	}
}

// runs job logic inside a goroutine
// uses context.WithTimeout to kill long-running jobs
func (s *Service) runJob(workerID int, job Job) {
	log.Printf("[Worker %d] Running job ID %d: %s", workerID, job.ID, job.Title)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan bool)

	go func() {
		time.Sleep(3 * time.Second) // simulate work
		done <- true
	}()

	select {
	case <-ctx.Done():
		s.updateJobStatus(job.ID, "failed")
		log.Printf("[Worker %d] Job ID %d timed out", workerID, job.ID)
	case <-done:
		s.updateJobStatus(job.ID, "done")
		log.Printf("[Worker %d] Completed job ID %d", workerID, job.ID)
	}
}

// locks access to shared state (s.jobs)
// updates job status safely from multiple workers
func (s *Service) updateJobStatus(id int, status string) {
	s.mu.Lock()

	for i := range s.jobs {
		if s.jobs[i].ID == id {
			s.jobs[i].Status = status
			s.jobs[i].UpdatedAt = time.Now()
			SaveJobs(s.jobs)
			break
		}
	}
}
