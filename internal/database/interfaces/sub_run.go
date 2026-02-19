package interfaces

import (
	"context"
	"time"

	subModel "github.com/bestruirui/bestsub/internal/models/sub"
)

// SubRunRepository 订阅运行日志数据访问接口
type SubRunRepository interface {
	CreateRun(ctx context.Context, run *subModel.RunLog) error
	UpdateRun(ctx context.Context, run *subModel.RunLog) error
	CreateEvent(ctx context.Context, event *subModel.RunEvent) error
	ListRuns(ctx context.Context, subID uint16, limit int) ([]subModel.RunLog, error)
	ListEvents(ctx context.Context, runID uint64, subID uint16) ([]subModel.RunEvent, error)
	CleanupBefore(ctx context.Context, before time.Time) error
}
