package job

import "time"

type Job struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Priority    int       `json:"priority"` // 1 = high, 5 = low
	Retries     int       `json:"retries"`
}
