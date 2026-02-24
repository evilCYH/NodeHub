package task

import (
	"errors"
	"time"

	"github.com/evilCYH/NodeHub/internal/database/op"
	"github.com/evilCYH/NodeHub/internal/models/setting"
	"github.com/panjf2000/ants/v2"
)

var pool *ants.Pool
var thread int

func Init(maxThread int) {
	if maxThread <= 0 {
		maxThread = 1
	}
	pool, _ = ants.NewPool(maxThread, ants.WithNonblocking(true))
	thread = maxThread
}

func Submit(fn func()) error {
	if pool == nil {
		return errors.New("task pool is not initialized")
	}
	maxRetry := op.GetSettingInt(setting.TASK_MAX_RETRY)
	if maxRetry < 0 {
		maxRetry = 0
	}
	maxTimeoutSec := op.GetSettingInt(setting.TASK_MAX_TIMEOUT)
	var deadline time.Time
	if maxTimeoutSec > 0 {
		deadline = time.Now().Add(time.Duration(maxTimeoutSec) * time.Second)
	}

	retriesLeft := maxRetry
	for {
		err := pool.Submit(fn)
		if err == nil {
			return nil
		}
		if !errors.Is(err, ants.ErrPoolOverload) {
			return err
		}
		if retriesLeft <= 0 {
			return err
		}
		if !deadline.IsZero() && time.Now().After(deadline) {
			return err
		}
		retriesLeft--
		time.Sleep(50 * time.Millisecond)
	}
}

func Release() {
	pool.Release()
}

func MaxThread() int {
	return thread
}
