package interfaces

import (
	"context"
	"time"

	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
)

// NodeLogRepository 节点日志数据访问接口
type NodeLogRepository interface {
	Create(ctx context.Context, log *nodeModel.NodeLog) error
	Query(ctx context.Context, query nodeModel.NodeLogQuery) ([]nodeModel.NodeLog, int64, error)
	CleanupBefore(ctx context.Context, before time.Time) error
}
