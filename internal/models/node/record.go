package node

import "time"

type Record struct {
	Base            Base       `json:"base"`
	Info            *Info      `json:"info,omitempty"`
	InitStatus      InitStatus `json:"init_status"`
	LastCheckAt     time.Time  `json:"last_check_at,omitempty"`
	LastCheckSource string     `json:"last_check_source,omitempty"`
	LastFailReason  string     `json:"last_fail_reason,omitempty"`
	FirstSeenAt     time.Time  `json:"first_seen_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
