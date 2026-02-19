package node

import "time"

type UpdateLog struct {
	ID         uint64    `json:"id"`
	SubID      uint16    `json:"sub_id"`
	CreatedAt  time.Time `json:"created_at"`
	DurationMs uint16    `json:"duration_ms"`
	RawCount   uint16    `json:"raw_count"`
	Candidate  uint16    `json:"candidate"`
	Duplicate  uint16    `json:"duplicate"`
	Invalid    uint16    `json:"invalid"`
	TestFailed uint16    `json:"test_failed"`
	Accepted   uint16    `json:"accepted"`
	Merged     uint16    `json:"merged"`
	Dropped    uint16    `json:"dropped"`
	Details    []string  `json:"details,omitempty"`
}

type UpdateLogResponse struct {
	Latest  *UpdateLog  `json:"latest"`
	History []UpdateLog `json:"history"`
}
