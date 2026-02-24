package channel

import (
	"bytes"

	"github.com/evilCYH/NodeHub/internal/modules/register"
)

type WebHook struct {
	Url string `json:"url" name:"WebHook地址"`
}

func (e *WebHook) Init() error {
	return nil
}

func (e *WebHook) Send(title string, body *bytes.Buffer) error {
	return nil
}

func init() {
	register.Notify(&WebHook{})
}
