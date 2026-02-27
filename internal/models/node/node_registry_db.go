package node

import "time"

// RegistryInfoSnapshot 用于持久化 Info 字段，避免 Queue 的 json:"-" 丢失数据
type RegistryInfoSnapshot struct {
	DelaySamples     []uint16 `json:"delay_samples,omitempty"`
	SpeedUpSamples   []uint32 `json:"speed_up_samples,omitempty"`
	SpeedDownSamples []uint32 `json:"speed_down_samples,omitempty"`
	Risk             uint8    `json:"risk"`
	AliveStatus      uint64   `json:"alive_status"`
	IP               uint32   `json:"ip"`
	Country          string   `json:"country"`
}

// RegistryRecordDB 节点主表持久化结构
type RegistryRecordDB struct {
	SubID           uint16
	UniqueKey       uint64
	Raw             []byte
	Info            *RegistryInfoSnapshot
	InitStatus      InitStatus
	LastCheckAt     time.Time
	LastCheckSource string
	LastFailReason  string
	FirstSeenAt     time.Time
	UpdatedAt       time.Time
}
