package node

import "time"

type UpdateLog struct {
	ID         uint64    `json:"id"`
	SubID      uint16    `json:"sub_id"`
	RunID      uint64    `json:"run_id"`
	CreatedAt  time.Time `json:"created_at"`
	DurationMs uint32    `json:"duration_ms"`
	RawCount   uint32    `json:"raw_count"`
	Candidate  uint32    `json:"candidate"`
	Duplicate  uint32    `json:"duplicate"`
	Invalid    uint32    `json:"invalid"`
	TestFailed uint32    `json:"test_failed"`
	Accepted   uint32    `json:"accepted"`
	Merged     uint32    `json:"merged"`
	Dropped    uint32    `json:"dropped"`
	Details    []string  `json:"details,omitempty"`
}

type UpdateLogResponse struct {
	Latest  *UpdateLog  `json:"latest"`
	History []UpdateLog `json:"history"`
}
