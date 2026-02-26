package cron

import (
	"context"
	"testing"
)

func TestFetchRemoveClearsFuncWithoutSchedule(t *testing.T) {
	subID := uint16(65001)
	fetchFunc.Store(subID, cronFunc{fn: func() {}, cronExpr: "* * * * *"})
	fetchScheduled.Delete(subID)
	fetchRunning.Delete(subID)
	testingRunning.Delete(subID)

	if err := FetchRemove(subID); err != nil {
		t.Fatalf("FetchRemove returned error: %v", err)
	}

	if _, ok := fetchFunc.Load(subID); ok {
		t.Fatalf("expected fetchFunc cleared for sub %d", subID)
	}
}

func TestFetchIsRunning(t *testing.T) {
	subID := uint16(65002)
	defer fetchRunning.Delete(subID)
	defer testingRunning.Delete(subID)

	if FetchIsRunning(subID) {
		t.Fatalf("expected not running initially")
	}

	fetchRunning.Store(subID, func() {})
	if !FetchIsRunning(subID) {
		t.Fatalf("expected running when fetchRunning contains sub")
	}
	fetchRunning.Delete(subID)

	testingRunning.Store(subID, context.CancelFunc(func() {}))
	if !FetchIsRunning(subID) {
		t.Fatalf("expected running when testingRunning contains sub")
	}
}
