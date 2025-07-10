// job/service.go
package job

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Service struct {
	jobs   []Job
	mu     sync.Mutex
	nextID int
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
		jobs:   jobs,
		nextID: maxID + 1,
	}, nil
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

func (s *Service) CreateJob(title, desc string) (*Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := Job{
		ID:          s.nextID,
		Title:       title,
		Description: desc,
		Status:      "queued",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	s.jobs = append(s.jobs, job)
	s.nextID++

	if err := SaveJobs(s.jobs); err != nil {
		return nil, err
	}
	return &job, nil
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

func (s *Service) StartWorker() {
	go func() {
		for {
			s.processNextJob()
			time.Sleep(3 * time.Second) // poll every 3s
		}
	}()
}

func (s *Service) processNextJob() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, job := range s.jobs {
		if job.Status == "queued" {
			fmt.Printf("[Worker] Running job ID %d: %s\n", job.ID, job.Title)
			s.jobs[i].Status = "running"
			s.jobs[i].UpdatedAt = time.Now()
			SaveJobs(s.jobs)

			// simulate work
			s.mu.Unlock() // release during sleep
			time.Sleep(5 * time.Second)
			s.mu.Lock()

			s.jobs[i].Status = "done"
			s.jobs[i].UpdatedAt = time.Now()
			SaveJobs(s.jobs)

			fmt.Printf("[Worker] Completed job ID %d: %s\n", job.ID, job.Title)
			return // process one job at a time
		}
	}
}
