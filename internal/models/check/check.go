package check

import (
	"context"
	"encoding/json"
	"time"

	"github.com/evilCYH/NodeHub/internal/utils/log"
)

type Instance interface {
	Init() error
	Run(ctx context.Context, log *log.Logger, subID []uint16) Result
}

type Data struct {
	ID     uint16 `db:"id" json:"id"`
	Name   string `db:"name" json:"name" description:"检测任务名称"`
	Enable bool   `db:"enable" json:"enable" description:"是否启用"`
	Task   string `db:"task" json:"task" description:"任务配置"`
	Config string `db:"config" json:"config" description:"检测器配置"`
	Result string `db:"result" json:"result" description:"检测结果"`
}

type Task struct {
	SubIdExclude  bool     `json:"sub_id_exclude" example:"false" description:"是否排除订阅ID"`
	SubID         []uint16 `json:"sub_id" example:"1" description:"订阅ID"`
	CronExpr      string   `json:"cron_expr" example:"0 0 * * *" description:"cron表达式"`
	Notify        bool     `json:"notify" example:"true" description:"是否通知"`
	NotifyChannel int      `json:"notify_channel" example:"1" description:"通知渠道"`
	LogWriteFile  bool     `json:"log_write_file" example:"true" description:"是否写入日志文件"`
	LogLevel      string   `json:"log_level" example:"info" description:"日志级别"`
	Timeout       int      `json:"timeout" example:"60" description:"超时时间 分钟"`
	Type          string   `json:"type" example:"test" description:"任务类型"`
}

type Result struct {
	Msg      string    `json:"msg" description:"消息"`
	Extra    any       `json:"extra" description:"额外信息"`
	LastRun  time.Time `json:"last_run" description:"上次运行时间"`
	Duration int64     `json:"duration" description:"运行时长(单位:毫秒)"`
}

type Request struct {
	Name   string `db:"name" json:"name" example:"测试检测任务" description:"检测任务名称"`
	Enable bool   `db:"enable" json:"enable" description:"是否启用"`
	Task   Task   `db:"task" json:"task" description:"任务配置"`
	Config any    `db:"config" json:"config" description:"检测器配置"`
}

type Response struct {
	ID     uint16 `db:"id" json:"id" description:"检测任务ID"`
	Name   string `db:"name" json:"name" description:"检测任务名称"`
	Enable bool   `db:"enable" json:"enable" description:"是否启用"`
	Task   Task   `db:"task" json:"task" description:"任务配置"`
	Config any    `db:"config" json:"config" description:"检测器配置"`
	Status string `db:"-" json:"status" description:"检测状态"`
	Result Result `db:"result" json:"result" description:"检测结果"`
}

func (r *Data) GenResponse(status string) Response {
	var resp Response
	resp.ID = r.ID
	resp.Name = r.Name
	resp.Enable = r.Enable
	resp.Status = status
	if err := json.Unmarshal([]byte(r.Task), &resp.Task); err != nil {
		return resp
	}
	if err := json.Unmarshal([]byte(r.Config), &resp.Config); err != nil {
		return resp
	}
	if err := json.Unmarshal([]byte(r.Result), &resp.Result); err != nil {
		return resp
	}
	return resp
}

func (r *Request) GenData() Data {
	var data Data
	taskBytes, err := json.Marshal(r.Task)
	if err != nil {
		log.Errorf("failed to marshal task: %v", err)
		return data
	}
	taskStr := string(taskBytes)
	configBytes, err := json.Marshal(r.Config)
	if err != nil {
		log.Errorf("failed to marshal config: %v", err)
		return data
	}
	configStr := string(configBytes)
	data.Task = taskStr
	data.Config = configStr
	data.Name = r.Name
	data.Enable = r.Enable
	return data
}
