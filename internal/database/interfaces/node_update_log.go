package interfaces

import (
	"context"
	"time"

	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
)

// NodeUpdateLogRepository 节点聚合统计日志数据访问接口
type NodeUpdateLogRepository interface {
	Create(ctx context.Context, log *nodeModel.UpdateLog) error
	List(ctx context.Context, subID uint16, limit int) ([]nodeModel.UpdateLog, error)
	CleanupBefore(ctx context.Context, before time.Time) error
}
