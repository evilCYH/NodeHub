package interfaces

import (
	"context"

	"github.com/evilCYH/NodeHub/internal/models/setting"
)

type SettingRepository interface {
	Create(ctx context.Context, setting *[]setting.Setting) error

	GetAll(ctx context.Context) (*[]setting.Setting, error)

	GetByKey(ctx context.Context, key []string) (*[]setting.Setting, error)

	Update(ctx context.Context, data *[]setting.Setting) error
}
