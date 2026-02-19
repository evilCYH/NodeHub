package op

import (
	"context"
	"time"

	nodeModel "github.com/bestruirui/bestsub/internal/models/node"
)

func CreateNodeLog(ctx context.Context, logEntry *nodeModel.NodeLog) error {
	return NodeLogRepo().Create(ctx, logEntry)
}

func QueryNodeLogs(ctx context.Context, query nodeModel.NodeLogQuery) ([]nodeModel.NodeLog, int64, error) {
	return NodeLogRepo().Query(ctx, query)
}

func CleanupNodeLogs(ctx context.Context, before time.Time) error {
	return NodeLogRepo().CleanupBefore(ctx, before)
}
