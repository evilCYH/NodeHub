package node

import "time"

type LogSource string

const (
	LogSourceInit    LogSource = "init"
	LogSourceCheck   LogSource = "check"
	LogSourceUnknown LogSource = "unknown"
)

// NodeLog 节点日志（初测 + 检测统一）
type NodeLog struct {
	ID        uint64    `json:"id"`
	SubID     uint16    `json:"sub_id"`
	NodeKey   uint64    `json:"node_key,string"`
	NodeName  string    `json:"node_name"`
	Level     string    `json:"level"`
	Source    LogSource `json:"source"`
	RunID     uint64    `json:"run_id"`
	CheckID   uint16    `json:"check_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type NodeLogQuery struct {
	SubID    uint16 `json:"sub_id"`
	Source   string `json:"source"`
	Level    string `json:"level"`
	Keyword  string `json:"keyword"`
	CheckID  uint16 `json:"check_id"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type NodeLogResponse struct {
	Total int64     `json:"total"`
	List  []NodeLog `json:"list"`
}
