package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
	subModel "github.com/evilCYH/NodeHub/internal/models/sub"
)

func TestSubDeleteCascadeRemovesRelatedData(t *testing.T) {
	repository, err := New(filepath.Join(t.TempDir(), "nodehub.db"))
	if err != nil {
		t.Fatalf("failed to create sqlite repo: %v", err)
	}
	t.Cleanup(func() {
		_ = repository.Close()
	})

	if err := repository.Migrate(); err != nil {
		t.Fatalf("failed to migrate sqlite repo: %v", err)
	}

	ctx := context.Background()
	subRepo := repository.Sub()
	subData := &subModel.Data{
		Enable:   true,
		Name:     "delete-cascade-sub",
		Tags:     "[]",
		CronExpr: "0 * * * *",
		Config:   `{"url":"https://example.com/sub"}`,
	}
	if err := subRepo.Create(ctx, subData); err != nil {
		t.Fatalf("failed to create sub: %v", err)
	}

	run := &subModel.RunLog{
		SubID:  subData.ID,
		Status: "success",
	}
	if err := repository.SubRun().CreateRun(ctx, run); err != nil {
		t.Fatalf("failed to create sub run: %v", err)
	}
	if err := repository.SubRun().CreateEvent(ctx, &subModel.RunEvent{
		RunID:   run.ID,
		SubID:   subData.ID,
		Step:    "fetch",
		Level:   "info",
		Message: "ok",
	}); err != nil {
		t.Fatalf("failed to create sub run event: %v", err)
	}

	if err := repository.NodeLog().Create(ctx, &nodeModel.NodeLog{
		SubID:    subData.ID,
		NodeKey:  12345,
		NodeName: "node-a",
		Level:    "info",
		Source:   nodeModel.LogSourceInit,
		Message:  "init pass",
	}); err != nil {
		t.Fatalf("failed to create node log: %v", err)
	}

	if err := repository.NodeUpdateLog().Create(ctx, &nodeModel.UpdateLog{
		SubID:    subData.ID,
		RawCount: 1,
		Accepted: 1,
	}); err != nil {
		t.Fatalf("failed to create node update log: %v", err)
	}

	if err := repository.NodeRegistry().Upsert(ctx, &nodeModel.RegistryRecordDB{
		SubID:       subData.ID,
		UniqueKey:   12345,
		Raw:         []byte("name: node-a\ntype: ss\n"),
		InitStatus:  nodeModel.InitPassed,
		FirstSeenAt: time.Now(),
		UpdatedAt:   time.Now(),
	}); err != nil {
		t.Fatalf("failed to create node registry: %v", err)
	}

	deleted, err := subRepo.DeleteCascade(ctx, subData.ID)
	if err != nil {
		t.Fatalf("DeleteCascade returned error: %v", err)
	}
	if !deleted {
		t.Fatalf("expected deleted=true")
	}

	gotSub, err := subRepo.GetByID(ctx, subData.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if gotSub != nil {
		t.Fatalf("expected sub removed, got %+v", gotSub)
	}

	runs, err := repository.SubRun().ListRuns(ctx, subData.ID, 10)
	if err != nil {
		t.Fatalf("ListRuns returned error: %v", err)
	}
	if len(runs) != 0 {
		t.Fatalf("expected no sub runs, got %d", len(runs))
	}

	events, err := repository.SubRun().ListEvents(ctx, run.ID, subData.ID)
	if err != nil {
		t.Fatalf("ListEvents returned error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no sub run events, got %d", len(events))
	}

	nodeLogs, total, err := repository.NodeLog().Query(ctx, nodeModel.NodeLogQuery{
		SubID:    subData.ID,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Query node logs returned error: %v", err)
	}
	if total != 0 || len(nodeLogs) != 0 {
		t.Fatalf("expected no node logs, total=%d len=%d", total, len(nodeLogs))
	}

	updateLogs, err := repository.NodeUpdateLog().List(ctx, subData.ID, 10)
	if err != nil {
		t.Fatalf("List node update logs returned error: %v", err)
	}
	if len(updateLogs) != 0 {
		t.Fatalf("expected no node update logs, got %d", len(updateLogs))
	}

	registryRows, err := repository.NodeRegistry().List(ctx)
	if err != nil {
		t.Fatalf("List node registry returned error: %v", err)
	}
	for _, row := range registryRows {
		if row.SubID == subData.ID {
			t.Fatalf("expected no node registry rows for sub %d", subData.ID)
		}
	}

	deleted, err = subRepo.DeleteCascade(ctx, subData.ID)
	if err != nil {
		t.Fatalf("DeleteCascade second call returned error: %v", err)
	}
	if deleted {
		t.Fatalf("expected deleted=false on second delete")
	}
}
