package cron

import (
	"time"

	"github.com/robfig/cron/v3"
)

type cronFunc struct {
	fn       func()
	cronExpr string
}

var scheduler = cron.New(cron.WithLocation(time.Local))

const (
	RunningStatus   = "running"
	ScheduledStatus = "scheduled"
	PendingStatus   = "pending"
	DisabledStatus  = "disabled"
)

func Start() {
	scheduler.Start()
	startLogCleanup()
}

func Stop() {
	scheduler.Stop()
}

func startLogCleanup() {
	_, _ = scheduler.AddFunc("0 0 3 * * *", func() {
		cleanupLogs()
	})
}
