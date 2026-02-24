package interfaces

import (
	"context"

	"github.com/evilCYH/NodeHub/internal/models/storage"
)

// 存储配置数据访问接口
type StorageRepository interface {
	// Create 创建存储配置
	Create(ctx context.Context, config *storage.Data) error

	// GetByID 根据ID获取存储配置
	GetByID(ctx context.Context, id uint16) (*storage.Data, error)

	// Update 更新存储配置
	Update(ctx context.Context, config *storage.Data) error

	// Delete 删除存储配置
	Delete(ctx context.Context, id uint16) error

	// List 获取存储配置列表
	List(ctx context.Context) (*[]storage.Data, error)
}
