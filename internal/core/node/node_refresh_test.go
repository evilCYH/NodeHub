package node

import (
	"bytes"
	"context"
	"encoding/gob"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/evilCYH/NodeHub/internal/config"
	"github.com/evilCYH/NodeHub/internal/database"
	"github.com/evilCYH/NodeHub/internal/database/op"
	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
	"github.com/evilCYH/NodeHub/internal/utils/generic"
)

func makeQueue[T generic.Integer](vals ...T) generic.Queue[T] {
	q := generic.NewQueue[T](5)
	for _, v := range vals {
		q.Update(v)
	}
	return *q
}

func resetNodeState() {
	pool = nil
	nodeExist = NewExist(16)
	nodeProcess = NewExist(16)
	registry = newRegistryStore()

	for k := range subInfoMap {
		delete(subInfoMap, k)
	}
	for k := range countryInfoMap {
		delete(countryInfoMap, k)
	}
	for k := range subAggBuf {
		delete(subAggBuf, k)
	}
	for k := range countryAggBuf {
		delete(countryAggBuf, k)
	}
}

func makeNodeData(subID uint16, key uint64, delay uint16, up uint32, down uint32, risk uint8, country string) nodeModel.Data {
	return nodeModel.Data{
		Base: nodeModel.Base{
			SubId:     subID,
			UniqueKey: key,
		},
		Info: &nodeModel.Info{
			Delay:     makeQueue[uint16](delay),
			SpeedUp:   makeQueue[uint32](up),
			SpeedDown: makeQueue[uint32](down),
			Risk:      risk,
			Country:   country,
		},
	}
}

func setupConfigAndDB(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	if err := config.Load(configPath); err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	dbPath := filepath.Join(dir, "data", "nodehub.db")
	if err := database.Initialize("sqlite", dbPath); err != nil {
		t.Fatalf("init database failed: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})
}

func setupConfigOnly(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	if err := config.Load(configPath); err != nil {
		t.Fatalf("load config failed: %v", err)
	}
}

func TestFinalizeRefreshesInfoWhenNoValidNodes(t *testing.T) {
	setupConfigAndDB(t)
	resetNodeState()

	subID := uint16(101)
	pool = []nodeModel.Data{
		makeNodeData(subID, 1, 120, 10, 20, 3, "US"),
	}
	subInfoMap[subID] = nodeModel.SimpleInfo{}

	stats := newAddStats(subID, 1, 0)
	stats.Finalize()

	got := GetSubInfo(subID)
	if got.Count != 1 {
		t.Fatalf("expected count=1, got %d", got.Count)
	}
	if got.Delay != 120 {
		t.Fatalf("expected delay=120, got %d", got.Delay)
	}
	if got.SpeedUp != 10 {
		t.Fatalf("expected speed_up=10, got %d", got.SpeedUp)
	}
	if got.SpeedDown != 20 {
		t.Fatalf("expected speed_down=20, got %d", got.SpeedDown)
	}
	if got.Risk != 3 {
		t.Fatalf("expected risk=3, got %d", got.Risk)
	}
}

func TestInitNodePoolRefreshesInfoAfterRestore(t *testing.T) {
	setupConfigOnly(t)
	resetNodeState()

	subID := uint16(202)
	restorePool := []nodeModel.Data{
		makeNodeData(subID, 2, 80, 30, 40, 2, "JP"),
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(&restorePool); err != nil {
		t.Fatalf("encode pool failed: %v", err)
	}

	sessionFile := config.Base().Session.NodePath
	if err := os.MkdirAll(filepath.Dir(sessionFile), 0755); err != nil {
		t.Fatalf("create session dir failed: %v", err)
	}
	if err := os.WriteFile(sessionFile, buf.Bytes(), 0600); err != nil {
		t.Fatalf("write session file failed: %v", err)
	}

	InitNodePool(10)

	got := GetSubInfo(subID)
	if got.Count != 1 {
		t.Fatalf("expected count=1, got %d", got.Count)
	}
	if got.Delay != 80 {
		t.Fatalf("expected delay=80, got %d", got.Delay)
	}
}

func TestInitNodePoolWithoutSessionClearsStaleInfo(t *testing.T) {
	setupConfigOnly(t)
	resetNodeState()

	subID := uint16(303)
	subInfoMap[subID] = nodeModel.SimpleInfo{Count: 9, Delay: 999}

	sessionFile := config.Base().Session.NodePath
	_ = os.Remove(sessionFile)

	InitNodePool(10)

	got := GetSubInfo(subID)
	if got.Count != 0 || got.Delay != 0 {
		t.Fatalf("expected stale info cleared, got %+v", got)
	}
}

func TestInitNodePoolRestoresRegistryFromDB(t *testing.T) {
	setupConfigAndDB(t)
	resetNodeState()

	now := time.Now().Truncate(time.Second)
	err := op.UpsertNodeRegistry(context.Background(), &nodeModel.RegistryRecordDB{
		SubID:           404,
		UniqueKey:       9001,
		Raw:             []byte("name: db-restored\ntype: ss\n"),
		Info:            &nodeModel.RegistryInfoSnapshot{DelaySamples: []uint16{88}, SpeedUpSamples: []uint32{11}, SpeedDownSamples: []uint32{22}, AliveStatus: nodeModel.Alive, Country: "HK"},
		InitStatus:      nodeModel.InitPassed,
		LastCheckAt:     now,
		LastCheckSource: "alive_task",
		LastFailReason:  "timeout",
		FirstSeenAt:     now.Add(-time.Hour),
		UpdatedAt:       now,
	})
	if err != nil {
		t.Fatalf("failed to seed node registry: %v", err)
	}

	InitNodePool(10)

	records := GetRegistryAll()
	if len(records) != 1 {
		t.Fatalf("expected 1 registry record, got %d", len(records))
	}
	record := records[0]
	if record.Base.SubId != 404 || record.Base.UniqueKey != 9001 {
		t.Fatalf("unexpected registry key: %+v", record.Base)
	}
	if record.LastCheckSource != "alive_task" || record.LastFailReason != "timeout" {
		t.Fatalf("unexpected last check fields: %+v", record)
	}
	if record.Info == nil || record.Info.Delay.Average() != 88 {
		t.Fatalf("expected restored info delay=88, got %+v", record.Info)
	}

	poolNodes := GetAll()
	if len(poolNodes) != 1 {
		t.Fatalf("expected pool rebuilt from db registry, got %d", len(poolNodes))
	}
}

func TestRegistryInfoSnapshotRoundTrip(t *testing.T) {
	info := &nodeModel.Info{
		Delay:       makeQueue[uint16](10, 20, 30),
		SpeedUp:     makeQueue[uint32](100, 200),
		SpeedDown:   makeQueue[uint32](300, 400),
		Risk:        3,
		AliveStatus: nodeModel.Alive | nodeModel.Country,
		IP:          123,
		Country:     "SG",
	}

	snapshot := snapshotFromInfo(info)
	restored := infoFromSnapshot(snapshot)
	if restored == nil {
		t.Fatalf("expected restored info not nil")
	}
	if restored.Delay.Average() != info.Delay.Average() {
		t.Fatalf("delay average mismatch: got %d want %d", restored.Delay.Average(), info.Delay.Average())
	}
	if restored.SpeedUp.Average() != info.SpeedUp.Average() {
		t.Fatalf("speed up average mismatch: got %d want %d", restored.SpeedUp.Average(), info.SpeedUp.Average())
	}
	if restored.SpeedDown.Average() != info.SpeedDown.Average() {
		t.Fatalf("speed down average mismatch: got %d want %d", restored.SpeedDown.Average(), info.SpeedDown.Average())
	}
	if restored.Country != "SG" || restored.AliveStatus != info.AliveStatus || restored.Risk != info.Risk {
		t.Fatalf("restored info mismatch: %+v", restored)
	}
}
