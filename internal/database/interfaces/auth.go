package interfaces

import (
	"context"

	"github.com/evilCYH/NodeHub/internal/models/auth"
)

// 单用户认证数据访问接口
type AuthRepository interface {
	// 获取认证信息
	Get(ctx context.Context) (*auth.Data, error)

	// 更新用户名
	UpdateName(ctx context.Context, name string) error

	// 更新密码
	UpdatePassword(ctx context.Context, hashPassword string) error

	// 初始化认证信息（首次创建密码）
	Initialize(ctx context.Context, auth *auth.Data) error

	// 验证是否已初始化
	IsInitialized(ctx context.Context) (bool, error)
}
