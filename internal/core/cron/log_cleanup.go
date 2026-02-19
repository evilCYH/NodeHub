package cron

import (
	"context"
	"time"

	"github.com/bestruirui/bestsub/internal/database/op"
	"github.com/bestruirui/bestsub/internal/models/setting"
	"github.com/bestruirui/bestsub/internal/utils/log"
)

func cleanupLogs() {
	nodeKeepDays := op.GetSettingInt(setting.NODE_LOG_KEEP_DAYS)
	if nodeKeepDays > 0 {
		before := time.Now().AddDate(0, 0, -nodeKeepDays)
		if err := op.CleanupNodeLogs(context.Background(), before); err != nil {
			log.Warnf("failed to cleanup node logs: %v", err)
		}
		if err := op.CleanupNodeUpdateLogs(context.Background(), before); err != nil {
			log.Warnf("failed to cleanup node update logs: %v", err)
		}
	}

	subKeepDays := op.GetSettingInt(setting.SUB_LOG_KEEP_DAYS)
	if subKeepDays > 0 {
		before := time.Now().AddDate(0, 0, -subKeepDays)
		if err := op.CleanupSubRuns(context.Background(), before); err != nil {
			log.Warnf("failed to cleanup sub run logs: %v", err)
		}
	}
}
