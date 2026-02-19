package sub

import "time"

type RunLog struct {
	ID         uint64    `json:"id"`
	SubID      uint16    `json:"sub_id"`
	Status     string    `json:"status"`
	Message    string    `json:"message,omitempty"`
	RawCount   uint32    `json:"raw_count"`
	Accepted   uint32    `json:"accepted"`
	CreatedAt  time.Time `json:"created_at"`
	DurationMs uint32    `json:"duration_ms"`
}

type RunEvent struct {
	ID        uint64    `json:"id"`
	RunID     uint64    `json:"run_id"`
	SubID     uint16    `json:"sub_id"`
	RunTime   time.Time `json:"run_time"`
	Step      string    `json:"step"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type RunLogResponse struct {
	Runs   []RunLog   `json:"runs"`
	Events []RunEvent `json:"events"`
}
