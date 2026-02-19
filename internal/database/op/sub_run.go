package op

import (
	"context"
	"time"

	subModel "github.com/bestruirui/bestsub/internal/models/sub"
)

func CreateSubRun(ctx context.Context, run *subModel.RunLog) error {
	return SubRunRepo().CreateRun(ctx, run)
}

func UpdateSubRun(ctx context.Context, run *subModel.RunLog) error {
	return SubRunRepo().UpdateRun(ctx, run)
}

func CreateSubRunEvent(ctx context.Context, event *subModel.RunEvent) error {
	return SubRunRepo().CreateEvent(ctx, event)
}

func ListSubRuns(ctx context.Context, subID uint16, limit int) ([]subModel.RunLog, error) {
	return SubRunRepo().ListRuns(ctx, subID, limit)
}

func ListSubRunEvents(ctx context.Context, runID uint64, subID uint16) ([]subModel.RunEvent, error) {
	return SubRunRepo().ListEvents(ctx, runID, subID)
}

func CleanupSubRuns(ctx context.Context, before time.Time) error {
	return SubRunRepo().CleanupBefore(ctx, before)
}
