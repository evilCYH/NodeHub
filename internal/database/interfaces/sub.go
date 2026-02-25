package interfaces

import (
	"context"
	"time"

	"github.com/evilCYH/NodeHub/internal/models/sub"
)

// SubRepository 订阅链接数据访问接口
type SubRepository interface {
	// Create 创建链接
	Create(ctx context.Context, link *sub.Data) error

	// GetByID 根据ID获取链接
	GetByID(ctx context.Context, id uint16) (*sub.Data, error)

	// Update 更新链接
	Update(ctx context.Context, link *sub.Data) error

	// Delete 删除链接
	Delete(ctx context.Context, id uint16) error

	// List 获取订阅链接列表
	List(ctx context.Context) (*[]sub.Data, error)

	// BatchCreate 批量创建订阅链接
	BatchCreate(ctx context.Context, links []*sub.Data) error

	// UpdateSubInfo 更新订阅流量与到期信息
	UpdateSubInfo(ctx context.Context, id uint16, upload, download, total, expire int64, infoUpdatedAt *time.Time) error
}
