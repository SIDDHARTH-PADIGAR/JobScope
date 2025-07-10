package job

import (
	"encoding/json"
	"io"
	"os"
	"sync"
)

const filePath = "data/jobs.json"

var mu sync.Mutex

func LoadJobs() ([]Job, error) {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Job{}, nil //empty DB
		}
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var jobs []Job
	if err := json.Unmarshal(bytes, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

func SaveJobs(jobs []Job) error {
	mu.Lock()
	defer mu.Unlock()

	bytes, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, bytes, 0644)
}
