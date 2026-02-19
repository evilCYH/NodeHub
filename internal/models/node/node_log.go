package node

import "time"

// NodeTestLog 节点级详细测试日志
type NodeTestLog struct {
	ID        uint64    `json:"id"`         // 自增ID
	SubID     uint16    `json:"sub_id"`     // 订阅ID
	NodeName  string    `json:"node_name"`  // 节点名称
	Level     string    `json:"level"`      // info / warn / error
	Message   string    `json:"message"`    // 具体日志内容
	CreatedAt time.Time `json:"created_at"` // 时间
}

// NodeTestLogQuery 节点日志查询参数
type NodeTestLogQuery struct {
	SubID    uint16 `json:"sub_id" form:"sub_id"`       // 订阅ID
	Level    string `json:"level" form:"level"`         // 筛选级别
	Keyword  string `json:"keyword" form:"keyword"`     // 搜索关键词
	Page     int    `json:"page" form:"page"`           // 页码
	PageSize int    `json:"page_size" form:"page_size"` // 每页数量
}

// NodeTestLogResponse 节点日志响应
type NodeTestLogResponse struct {
	Total int64         `json:"total"`
	List  []NodeTestLog `json:"list"`
}
