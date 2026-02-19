package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bestruirui/bestsub/internal/core/check"
	"github.com/bestruirui/bestsub/internal/core/node"
	"github.com/bestruirui/bestsub/internal/database/op"
	checkModel "github.com/bestruirui/bestsub/internal/models/check"
	"github.com/bestruirui/bestsub/internal/utils/generic"
	"github.com/bestruirui/bestsub/internal/utils/log"
	"github.com/robfig/cron/v3"
)

var checkFunc = generic.MapOf[uint16, cronFunc]{}
var checkScheduled = generic.MapOf[uint16, cron.EntryID]{}
var checkRunning = generic.MapOf[uint16, context.CancelFunc]{}

func CheckLoad() {
	checkData, err := op.GetCheckList()
	if err != nil {
		log.Errorf("failed to load sub data: %v", err)
		return
	}
	for _, data := range checkData {
		CheckAdd(&data)
	}
}
func CheckAdd(data *checkModel.Data) error {
	var taskConfig checkModel.Task
	if err := json.Unmarshal([]byte(data.Task), &taskConfig); err != nil {
		log.Errorf("failed to unmarshal task config: %v", err)
		return err
	}
	checkFunc.Store(data.ID, cronFunc{
		fn: func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(taskConfig.Timeout)*time.Minute)
			ctx = context.WithValue(ctx, "check_id", data.ID)
			checkRunning.Store(data.ID, cancel)
			defer func() {
				cancel()
				checkRunning.Delete(data.ID)
			}()
			logger, err := log.NewTaskLogger("check", data.ID, taskConfig.LogLevel, taskConfig.LogWriteFile)
			if err != nil {
				log.Errorf("failed to create logger: %v", err)
				return
			}
			go func() {
				<-ctx.Done()
				logger.Close()
			}()
			checker, err := check.Get(taskConfig.Type, data.Config)
			if err != nil {
				log.Errorf("failed to get execer: %v", err)
				return
			}
			log.Infof("%s task %d start", taskConfig.Type, data.ID)
			var result checkModel.Result
			if taskConfig.SubIdExclude {
				result = checker.Run(ctx, logger, node.GetBySubIdExclude(taskConfig.SubID))
			} else {
				result = checker.Run(ctx, logger, taskConfig.SubID)
			}
			log.Infof("%s task %d end", taskConfig.Type, data.ID)
			op.UpdateCheckResult(data.ID, result)
			node.RefreshInfo()
		},
		cronExpr: taskConfig.CronExpr,
	})
	if data.Enable {
		CheckEnable(data.ID)
	}
	return nil
}
func CheckUpdate(data *checkModel.Data) error {
	CheckRemove(data.ID)
	CheckAdd(data)
	return nil
}

func CheckRun(id uint16) error {
	if ft, ok := checkFunc.Load(id); ok {
		go ft.fn()
		return nil
	} else {
		return fmt.Errorf("check task %d not found", id)
	}
}

func CheckEnable(id uint16) error {
	if _, ok := checkScheduled.Load(id); ok {
		log.Warnf("check task %d already scheduled", id)
		return nil
	}
	if ft, ok := checkFunc.Load(id); ok {
		entryID, err := scheduler.AddFunc(ft.cronExpr, ft.fn)
		if err != nil {
			log.Errorf("failed to add task: %v", err)
			return err
		}
		checkScheduled.Store(id, entryID)
	}
	return nil
}
func CheckDisable(id uint16) error {
	if entryID, ok := checkScheduled.Load(id); ok {
		scheduler.Remove(entryID)
		checkScheduled.Delete(id)
		if cancel, ok := checkRunning.Load(id); ok {
			cancel()
			checkRunning.Delete(id)
		}
	}
	return nil
}

func CheckRemove(id uint16) error {
	if entryID, ok := checkScheduled.Load(id); ok {
		scheduler.Remove(entryID)
		checkScheduled.Delete(id)
		checkFunc.Delete(id)
		if cancel, ok := checkRunning.Load(id); ok {
			cancel()
			checkRunning.Delete(id)
		}
	}
	return nil
}
func CheckStop(id uint16) error {
	if cancel, ok := checkRunning.Load(id); ok {
		cancel()
		checkRunning.Delete(id)
	}
	return nil
}
func CheckStatus(id uint16) string {
	if _, ok := checkRunning.Load(id); ok {
		return RunningStatus
	}
	if _, ok := checkScheduled.Load(id); ok {
		return ScheduledStatus
	}
	if _, ok := checkFunc.Load(id); ok {
		return PendingStatus
	}
	return DisabledStatus
}
