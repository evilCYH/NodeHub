package cron

import (
	"context"
	"time"

	"github.com/evilCYH/NodeHub/internal/database/op"
	"github.com/evilCYH/NodeHub/internal/models/setting"
	"github.com/evilCYH/NodeHub/internal/utils/log"
)

func cleanupLogs() {
	retentionDays := op.GetSettingInt(setting.LOG_RETENTION_DAYS)
	if retentionDays <= 0 {
		return
	}
	before := time.Now().AddDate(0, 0, -retentionDays)
	if err := op.CleanupNodeLogs(context.Background(), before); err != nil {
		log.Warnf("failed to cleanup node logs: %v", err)
	}
	if err := op.CleanupNodeUpdateLogs(context.Background(), before); err != nil {
		log.Warnf("failed to cleanup node update logs: %v", err)
	}
	if err := op.CleanupSubRuns(context.Background(), before); err != nil {
		log.Warnf("failed to cleanup sub run logs: %v", err)
	}
	if err := log.CleanupOldLogs(retentionDays); err != nil {
		log.Warnf("failed to cleanup file logs: %v", err)
	}
}
