package cron

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/evilCYH/NodeHub/internal/core/fetch"
	"github.com/evilCYH/NodeHub/internal/database/op"
	subModel "github.com/evilCYH/NodeHub/internal/models/sub"
	"github.com/evilCYH/NodeHub/internal/utils/generic"
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"github.com/robfig/cron/v3"
)

var fetchFunc = generic.MapOf[uint16, cronFunc]{}
var fetchScheduled = generic.MapOf[uint16, cron.EntryID]{}
var fetchRunning = generic.MapOf[uint16, context.CancelFunc]{}
var testingRunning = generic.MapOf[uint16, context.CancelFunc]{}

func FetchLoad() {
	subData, err := op.GetSubList(context.Background())
	if err != nil {
		log.Errorf("failed to load sub data: %v", err)
		return
	}
	for i := range subData {
		FetchAdd(&subData[i])
	}
}

func FetchAdd(data *subModel.Data) error {
	subID := data.ID
	cronExpr := data.CronExpr
	config := data.Config
	enable := data.Enable

	fetchFunc.Store(subID, cronFunc{
		fn: func() {
			runFetch(subID, config)
		},
		cronExpr: cronExpr,
	})
	if enable {
		FetchEnable(subID)
	}
	return nil
}

func runFetch(subID uint16, config string) subModel.Result {
	fetchCtx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	fetchRunning.Store(subID, cancel)
	defer func() {
		cancel()
		fetchRunning.Delete(subID)
	}()
	if _, err := op.GetSubByID(context.Background(), subID); err != nil {
		log.Warnf("fetch task %d skipped: %v", subID, err)
		msg := "fetch task not found"
		if !errors.Is(err, op.ErrSubNotFound) {
			msg = "fetch task unavailable"
		}
		return subModel.Result{
			Msg:     msg,
			LastRun: time.Now(),
		}
	}

	result := fetch.Do(fetchCtx, subID, config)

	// 阶段2：异步等待初测完成，保持running状态
	if done, ok := fetch.GetTestingDone(subID); ok {
		testCtx, testCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		testingRunning.Store(subID, testCancel)

		go func() {
			defer func() {
				testCancel()
				testingRunning.Delete(subID)
				fetch.DeleteTestingDone(subID)
			}()

			select {
			case <-done:
				log.Infof("sub %d testing completed", subID)
			case <-testCtx.Done():
				log.Warnf("sub %d testing timeout or cancelled", subID)
			}
		}()
	}

	updateCtx, updateCancel := context.WithTimeout(context.Background(), 2*time.Second)
	if err := op.UpdateSubResult(updateCtx, subID, result); err != nil {
		log.Warnf("failed to update sub result: %v", err)
	}
	updateCancel()
	sub, err := op.GetSubByID(context.Background(), subID)
	if err != nil {
		log.Warnf("failed to get sub by id: %v", err)
		return result
	}
	if !sub.Enable {
		FetchDisable(subID)
		log.Infof("fetch task %d auto disable", subID)
	}
	return result
}

func FetchRun(subID uint16) subModel.Result {
	if _, ok := fetchFunc.Load(subID); !ok {
		log.Warnf("fetch task %d not found", subID)
		return subModel.Result{
			Msg:     "fetch task not found",
			LastRun: time.Now(),
		}
	}
	sub, err := op.GetSubByID(context.Background(), subID)
	if err != nil {
		log.Warnf("failed to get sub by id: %v", err)
		return subModel.Result{
			Msg:     "fetch task not found",
			LastRun: time.Now(),
		}
	}
	return runFetch(subID, sub.Config)
}

func FetchEnable(subID uint16) error {
	if _, ok := fetchScheduled.Load(subID); ok {
		log.Warnf("fetch task %d already scheduled", subID)
		return nil
	}
	if ft, ok := fetchFunc.Load(subID); ok {
		entryID, err := scheduler.AddFunc(ft.cronExpr,
			func() {
				time.Sleep(time.Duration(rand.Intn(100)) * time.Second)
				ft.fn()
			})
		if err != nil {
			log.Errorf("failed to add task: %v", err)
			return err
		}
		fetchScheduled.Store(subID, entryID)
	}
	return nil
}
func FetchDisable(subID uint16) error {
	if entryID, ok := fetchScheduled.Load(subID); ok {
		scheduler.Remove(entryID)
		fetchScheduled.Delete(subID)
		if cancel, ok := fetchRunning.Load(subID); ok {
			cancel()
			fetchRunning.Delete(subID)
		}
		// 停止初测
		if testCancel, ok := testingRunning.Load(subID); ok {
			testCancel()
			testingRunning.Delete(subID)
			fetch.DeleteTestingDone(subID)
		}
	}
	return nil
}

func FetchRemove(subID uint16) error {
	if entryID, ok := fetchScheduled.Load(subID); ok {
		scheduler.Remove(entryID)
		fetchScheduled.Delete(subID)
	}
	fetchFunc.Delete(subID)
	if cancel, ok := fetchRunning.Load(subID); ok {
		cancel()
		fetchRunning.Delete(subID)
	}
	// 停止初测
	if testCancel, ok := testingRunning.Load(subID); ok {
		testCancel()
		testingRunning.Delete(subID)
		fetch.DeleteTestingDone(subID)
	}
	return nil
}
func FetchStop(subID uint16) error {
	if cancel, ok := fetchRunning.Load(subID); ok {
		cancel()
		fetchRunning.Delete(subID)
	}
	// 停止初测
	if testCancel, ok := testingRunning.Load(subID); ok {
		testCancel()
		testingRunning.Delete(subID)
		fetch.DeleteTestingDone(subID)
	}
	return nil
}
func FetchUpdate(data *subModel.Data) error {
	FetchRemove(data.ID)
	FetchAdd(data)
	return nil
}
func FetchStatus(subID uint16, enable bool) string {
	if _, ok := fetchRunning.Load(subID); ok {
		return RunningStatus
	}
	if _, ok := testingRunning.Load(subID); ok {
		return RunningStatus // 初测中也返回running
	}
	if !enable {
		return DisabledStatus
	}
	if _, ok := fetchScheduled.Load(subID); ok {
		return ScheduledStatus
	}
	if _, ok := fetchFunc.Load(subID); ok {
		return PendingStatus
	}
	return DisabledStatus
}

func FetchIsRunning(subID uint16) bool {
	if _, ok := fetchRunning.Load(subID); ok {
		return true
	}
	if _, ok := testingRunning.Load(subID); ok {
		return true
	}
	return false
}
