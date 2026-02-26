package sub

import (
	"encoding/json"
	"time"

	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
)

type Data struct {
	ID            uint16     `db:"id" json:"id"`
	SortOrder     int        `db:"sort_order" json:"sort_order"`
	Enable        bool       `db:"enable" json:"enable"`
	Name          string     `db:"name" json:"name"`
	Tags          string     `db:"tags" json:"tags"`
	CronExpr      string     `db:"cron_expr" json:"cron_expr"`
	Config        string     `db:"config" json:"config"`
	Result        string     `db:"result" json:"result"`
	Upload        int64      `db:"upload" json:"upload"`
	Download      int64      `db:"download" json:"download"`
	Total         int64      `db:"total" json:"total"`
	Expire        int64      `db:"expire" json:"expire"`
	InfoUpdatedAt *time.Time `db:"info_updated_at" json:"info_updated_at"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

type Config struct {
	Url                  string   `json:"url"`
	Proxy                bool     `json:"proxy"`
	Timeout              int      `json:"timeout"`
	ProtocolFilterEnable bool     `json:"protocol_filter_enable"`
	ProtocolFilterMode   bool     `json:"protocol_filter_mode"`
	ProtocolFilter       []string `json:"protocol_filter"`
}

type Result struct {
	Success       uint32    `json:"success,omitempty" description:"成功次数"`
	Fail          uint32    `json:"fail,omitempty" description:"失败次数"`
	NodeNullCount uint32    `json:"node_null_count,omitempty" description:"节点为空次数"`
	Msg           string    `json:"msg,omitempty" description:"消息"`
	LastStatus    string    `json:"last_status,omitempty" description:"上次运行状态"`
	RawCount      uint32    `json:"raw_count,omitempty" description:"节点数量"`
	LastRun       time.Time `json:"last_run,omitempty" description:"上次运行时间"`
	Duration      uint32    `json:"duration,omitempty" description:"运行时长(单位:毫秒)"`
}

type NameAndID struct {
	ID   uint16 `json:"id"`
	Name string `json:"name"`
}

type SortOrderItem struct {
	ID        uint16 `json:"id" binding:"required"`
	SortOrder int    `json:"sort_order" binding:"required"`
}

type UpdateOrderRequest struct {
	Orders []SortOrderItem `json:"orders" binding:"required"`
}

type Request struct {
	Name     string   `json:"name" description:"订阅任务名称"`
	Tags     []string `json:"tags" description:"订阅标签"`
	Enable   bool     `json:"enable" description:"是否启用"`
	CronExpr string   `json:"cron_expr" example:"0 0 * * *" description:"cron表达式"`
	Config   Config   `json:"config"`
}

type Response struct {
	ID            uint16               `json:"id" description:"订阅任务ID"`
	Name          string               `json:"name" description:"订阅任务名称"`
	Tags          []string             `json:"tags" description:"订阅标签"`
	Enable        bool                 `json:"enable" description:"是否启用"`
	CronExpr      string               `json:"cron_expr" description:"cron表达式"`
	Config        Config               `json:"config" description:"订阅器配置"`
	Status        string               `json:"status" description:"订阅状态"`
	Result        Result               `json:"result" description:"订阅结果"`
	Info          nodeModel.SimpleInfo `json:"info" description:"订阅信息"`
	Upload        int64                `json:"upload" description:"已上传流量(字节)"`
	Download      int64                `json:"download" description:"已下载流量(字节)"`
	Total         int64                `json:"total" description:"总流量(字节)"`
	Expire        int64                `json:"expire" description:"到期时间(Unix秒)"`
	InfoUpdatedAt *time.Time           `json:"info_updated_at" description:"流量信息更新时间"`
	CreatedAt     time.Time            `json:"created_at" description:"创建时间"`
	UpdatedAt     time.Time            `json:"updated_at" description:"更新时间"`
}

func (c *Request) GenData(id uint16) Data {
	configBytes, err := json.Marshal(c.Config)
	if err != nil {
		return Data{}
	}
	tags, _ := json.Marshal(c.Tags)
	return Data{
		ID:       id,
		Name:     c.Name,
		Tags:     string(tags),
		Enable:   c.Enable,
		CronExpr: c.CronExpr,
		Config:   string(configBytes),
	}
}
func (d *Data) GenResponse(status string, subInfo nodeModel.SimpleInfo) Response {
	var config Config
	json.Unmarshal([]byte(d.Config), &config)
	var result Result
	json.Unmarshal([]byte(d.Result), &result)
	tags := make([]string, 0)
	json.Unmarshal([]byte(d.Tags), &tags)
	return Response{
		ID:            d.ID,
		Name:          d.Name,
		Tags:          tags,
		Enable:        d.Enable,
		CronExpr:      d.CronExpr,
		Config:        config,
		Status:        status,
		Result:        result,
		Info:          subInfo,
		Upload:        d.Upload,
		Download:      d.Download,
		Total:         d.Total,
		Expire:        d.Expire,
		InfoUpdatedAt: d.InfoUpdatedAt,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}
