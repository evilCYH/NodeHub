package channel

import (
	"context"

	"github.com/evilCYH/NodeHub/internal/modules/register"
)

func init() {
	register.Storage(&WebDAV{})
}

type WebDAV struct {
	url      string `json:"url" type:"string" required:"true" description:"WebDAV地址"`
	username string `json:"username" type:"string" required:"true" description:"WebDAV用户名"`
	password string `json:"password" type:"string" required:"true" description:"WebDAV密码"`
}

func (w *WebDAV) Init() error {
	return nil
}

func (w *WebDAV) Upload(ctx context.Context) error {
	return nil
}
