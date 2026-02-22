package notify

import (
	"bytes"
	"encoding/json"
)

type Data struct {
	ID     uint16 `db:"id" json:"id"`
	Name   string `db:"name" json:"name"`
	Type   string `db:"type" json:"type"`
	Config string `db:"config" json:"config"`
}
type NameAndID struct {
	ID   uint16 `json:"id"`
	Name string `json:"name"`
}
type Request struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Config any    `json:"config"`
}

type Response struct {
	ID     uint16 `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Config any    `json:"config"`
}

type Template struct {
	Type    string `db:"type" json:"type"`
	Title   string `db:"title" json:"title"`
	Content string `db:"content" json:"content"`
}

type Instance interface {
	Init() error
	Send(title string, body *bytes.Buffer) error
}

const (
	TypeLoginSuccess uint16 = 1 << 0 // 登录成功通知
	TypeLoginFailed  uint16 = 1 << 1 // 登录失败通知
)

var TypeMap = map[uint16]string{
	TypeLoginSuccess: "login_success",
	TypeLoginFailed:  "login_failed",
}

func (c *Request) GenData(id uint16) Data {
	configBytes, err := json.Marshal(c.Config)
	if err != nil {
		return Data{}
	}
	return Data{
		ID:     id,
		Name:   c.Name,
		Type:   c.Type,
		Config: string(configBytes),
	}
}
func (d *Data) GenResponse() Response {
	var config any
	json.Unmarshal([]byte(d.Config), &config)
	return Response{
		ID:     d.ID,
		Name:   d.Name,
		Type:   d.Type,
		Config: config,
	}
}
