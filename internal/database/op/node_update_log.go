package op

import (
	"context"
	"time"

	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
)

func CreateNodeUpdateLog(ctx context.Context, logEntry *nodeModel.UpdateLog) error {
	return NodeUpdateLogRepo().Create(ctx, logEntry)
}

func ListNodeUpdateLogs(ctx context.Context, subID uint16, limit int) ([]nodeModel.UpdateLog, error) {
	return NodeUpdateLogRepo().List(ctx, subID, limit)
}

func ListNodeUpdateLogsByRunID(ctx context.Context, subID uint16, runID uint64) ([]nodeModel.UpdateLog, error) {
	return NodeUpdateLogRepo().ListByRunID(ctx, subID, runID)
}

func CleanupNodeUpdateLogs(ctx context.Context, before time.Time) error {
	return NodeUpdateLogRepo().CleanupBefore(ctx, before)
}
