package node

import "time"

type Response struct {
	SubID           uint16     `json:"sub_id"`
	UniqueKey       uint64     `json:"unique_key"`
	UniqueKeyStr    string     `json:"unique_key_str,omitempty"`
	Name            string     `json:"name"`
	Type            string     `json:"type"`
	Reason          string     `json:"reason,omitempty"`
	Delay           uint16     `json:"delay"`
	SpeedUp         uint32     `json:"speed_up"`
	SpeedDown       uint32     `json:"speed_down"`
	Risk            uint8      `json:"risk"`
	AliveStatus     uint64     `json:"alive_status"`
	Country         string     `json:"country"`
	InitStatus      InitStatus `json:"init_status,omitempty"`
	LastCheckAt     time.Time  `json:"last_check_at,omitempty"`
	LastCheckSource string     `json:"last_check_source,omitempty"`
	LastFailReason  string     `json:"last_fail_reason,omitempty"`
}

type DetailResponse struct {
	Response
	Raw map[string]any `json:"raw"`
}
