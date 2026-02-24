package interfaces

import (
	"context"

	"github.com/evilCYH/NodeHub/internal/models/share"
)

// 分享链接数据访问接口
type ShareRepository interface {
	// Create 创建分享链接
	Create(ctx context.Context, shareLink *share.Data) error

	// GetByID 根据ID获取分享链接
	GetByID(ctx context.Context, id uint16) (*share.Data, error)

	// Update 更新分享链接
	Update(ctx context.Context, shareLink *share.Data) error

	// UpdateAccessCount 更新分享链接访问次数
	UpdateAccessCount(ctx context.Context, shareLink *[]share.UpdateAccessCountDB) error

	// GetConfigByToken 根据token获取分享链接配置
	GetGenByToken(ctx context.Context, token string) (string, error)

	// Delete 删除分享链接
	Delete(ctx context.Context, id uint16) error

	// List 获取分享链接列表
	List(ctx context.Context) (*[]share.Data, error)
}
