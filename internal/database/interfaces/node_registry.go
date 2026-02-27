package interfaces

import (
	"context"

	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
)

// NodeRegistryRepository 节点主表状态持久化接口
type NodeRegistryRepository interface {
	Upsert(ctx context.Context, record *nodeModel.RegistryRecordDB) error
	List(ctx context.Context) ([]nodeModel.RegistryRecordDB, error)
	DeleteBySubID(ctx context.Context, subID uint16) error
}
