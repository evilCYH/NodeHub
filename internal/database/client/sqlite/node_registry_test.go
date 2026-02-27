package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
)

func TestNodeRegistryUpsertListAndDelete(t *testing.T) {
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
	firstSeen := time.Now().Add(-time.Hour).Truncate(time.Second)
	updatedAt := time.Now().Truncate(time.Second)
	record := &nodeModel.RegistryRecordDB{
		SubID:           12,
		UniqueKey:       345,
		Raw:             []byte("name: node-a\ntype: ss\n"),
		Info:            &nodeModel.RegistryInfoSnapshot{DelaySamples: []uint16{100, 120}, SpeedUpSamples: []uint32{30}, SpeedDownSamples: []uint32{40}, Risk: 2, AliveStatus: nodeModel.Alive, Country: "US"},
		InitStatus:      nodeModel.InitPassed,
		LastCheckAt:     updatedAt,
		LastCheckSource: "alive_task",
		LastFailReason:  "",
		FirstSeenAt:     firstSeen,
		UpdatedAt:       updatedAt,
	}
	if err := repository.NodeRegistry().Upsert(ctx, record); err != nil {
		t.Fatalf("upsert node registry failed: %v", err)
	}

	record2 := &nodeModel.RegistryRecordDB{
		SubID:           12,
		UniqueKey:       345,
		Raw:             []byte("name: node-a\ntype: ss\n"),
		Info:            &nodeModel.RegistryInfoSnapshot{DelaySamples: []uint16{90}, SpeedUpSamples: []uint32{50}, SpeedDownSamples: []uint32{80}, Risk: 1, AliveStatus: nodeModel.Alive | nodeModel.Country, Country: "JP"},
		InitStatus:      nodeModel.InitPassed,
		LastCheckAt:     updatedAt.Add(time.Minute),
		LastCheckSource: "speed_task",
		LastFailReason:  "timeout",
		FirstSeenAt:     time.Now(),
		UpdatedAt:       updatedAt.Add(time.Minute),
	}
	if err := repository.NodeRegistry().Upsert(ctx, record2); err != nil {
		t.Fatalf("upsert node registry second time failed: %v", err)
	}

	list, err := repository.NodeRegistry().List(ctx)
	if err != nil {
		t.Fatalf("list node registry failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 node registry row, got %d", len(list))
	}
	got := list[0]
	if got.SubID != 12 || got.UniqueKey != 345 {
		t.Fatalf("unexpected key values: %+v", got)
	}
	if got.FirstSeenAt.IsZero() {
		t.Fatalf("expected first_seen_at set")
	}
	if got.LastCheckSource != "speed_task" || got.LastFailReason != "timeout" {
		t.Fatalf("unexpected last check fields: %+v", got)
	}
	if got.Info == nil || got.Info.Country != "JP" || len(got.Info.DelaySamples) != 1 {
		t.Fatalf("unexpected info snapshot: %+v", got.Info)
	}

	if err := repository.NodeRegistry().DeleteBySubID(ctx, 12); err != nil {
		t.Fatalf("delete node registry by sub_id failed: %v", err)
	}
	list, err = repository.NodeRegistry().List(ctx)
	if err != nil {
		t.Fatalf("list node registry after delete failed: %v", err)
	}
	for _, item := range list {
		if item.SubID == 12 {
			t.Fatalf("expected deleted sub_id rows")
		}
	}
}
