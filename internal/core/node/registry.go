package node

import (
	"context"
	"sync"
	"time"

	"github.com/evilCYH/NodeHub/internal/database/op"
	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
	"github.com/evilCYH/NodeHub/internal/utils/log"
)

type registryKey struct {
	subID     uint16
	uniqueKey uint64
}

type registryStore struct {
	mu    sync.RWMutex
	items map[registryKey]nodeModel.Record
}

func newRegistryStore() *registryStore {
	return &registryStore{items: make(map[registryKey]nodeModel.Record)}
}

func (r *registryStore) Upsert(record nodeModel.Record) {
	item := r.upsertNoPersist(record)
	r.persist(item)
}

func (r *registryStore) Replace(records []nodeModel.Record) {
	now := time.Now()
	newItems := make(map[registryKey]nodeModel.Record, len(records))
	for _, record := range records {
		if record.InitStatus == "" {
			record.InitStatus = nodeModel.InitUnknown
		}
		if record.FirstSeenAt.IsZero() {
			record.FirstSeenAt = now
		}
		if record.UpdatedAt.IsZero() {
			record.UpdatedAt = now
		}
		key := registryKey{subID: record.Base.SubId, uniqueKey: record.Base.UniqueKey}
		newItems[key] = record
	}
	r.mu.Lock()
	r.items = newItems
	r.mu.Unlock()
}

func (r *registryStore) upsertNoPersist(record nodeModel.Record) nodeModel.Record {
	now := time.Now()
	r.mu.Lock()
	key := registryKey{subID: record.Base.SubId, uniqueKey: record.Base.UniqueKey}
	if existing, ok := r.items[key]; ok {
		if record.Info == nil {
			record.Info = existing.Info
		}
		if record.InitStatus == nodeModel.InitUnknown {
			record.InitStatus = existing.InitStatus
		}
		if record.LastCheckAt.IsZero() {
			record.LastCheckAt = existing.LastCheckAt
		}
		if record.LastCheckSource == "" {
			record.LastCheckSource = existing.LastCheckSource
		}
		if record.LastFailReason == "" {
			record.LastFailReason = existing.LastFailReason
		}
		record.FirstSeenAt = existing.FirstSeenAt
		record.UpdatedAt = now
	} else {
		if record.InitStatus == "" {
			record.InitStatus = nodeModel.InitUnknown
		}
		record.FirstSeenAt = now
		record.UpdatedAt = now
	}
	r.items[key] = record
	persistCopy := copyRecord(record)
	r.mu.Unlock()
	return persistCopy
}

func (r *registryStore) persist(record nodeModel.Record) {
	if !op.HasRepo() {
		return
	}
	dbRecord := toDBRecord(record)
	if err := op.UpsertNodeRegistry(context.Background(), &dbRecord); err != nil {
		log.Warnf("failed to persist node registry sub_id=%d key=%d: %v", record.Base.SubId, record.Base.UniqueKey, err)
	}
}

func (r *registryStore) Get(subID uint16, uniqueKey uint64) (nodeModel.Record, bool) {
	r.mu.RLock()
	item, ok := r.items[registryKey{subID: subID, uniqueKey: uniqueKey}]
	r.mu.RUnlock()
	return item, ok
}

func (r *registryStore) UpdateInfo(subID uint16, uniqueKey uint64, info *nodeModel.Info, source string) {
	var persistRecord nodeModel.Record
	var shouldPersist bool
	r.mu.Lock()
	key := registryKey{subID: subID, uniqueKey: uniqueKey}
	item, ok := r.items[key]
	if ok {
		item.Info = info
		item.LastCheckAt = time.Now()
		item.LastCheckSource = source
		item.UpdatedAt = time.Now()
		r.items[key] = item
		persistRecord = copyRecord(item)
		shouldPersist = true
	}
	r.mu.Unlock()
	if shouldPersist {
		r.persist(persistRecord)
	}
}

func (r *registryStore) UpdateAlive(subID uint16, uniqueKey uint64, alive bool, delay uint16, source string) {
	var persistRecord nodeModel.Record
	var shouldPersist bool
	r.mu.Lock()
	key := registryKey{subID: subID, uniqueKey: uniqueKey}
	item, ok := r.items[key]
	if ok && item.Info != nil {
		item.Info.SetAliveStatus(nodeModel.Alive, alive)
		if alive && delay > 0 {
			item.Info.Delay.Update(delay)
		}
		if !alive && source == "alive_task" {
			item.LastFailReason = "alive_check_failed"
		}
		item.LastCheckAt = time.Now()
		item.LastCheckSource = source
		item.UpdatedAt = time.Now()
		r.items[key] = item
		persistRecord = copyRecord(item)
		shouldPersist = true
	}
	r.mu.Unlock()
	if shouldPersist {
		r.persist(persistRecord)
	}
}

func (r *registryStore) UpdateCountry(subID uint16, uniqueKey uint64, country string, ok bool, source string) {
	var persistRecord nodeModel.Record
	var shouldPersist bool
	r.mu.Lock()
	key := registryKey{subID: subID, uniqueKey: uniqueKey}
	item, exists := r.items[key]
	if exists && item.Info != nil {
		item.Info.Country = country
		item.Info.SetAliveStatus(nodeModel.Country, ok)
		item.LastCheckAt = time.Now()
		item.LastCheckSource = source
		item.UpdatedAt = time.Now()
		r.items[key] = item
		persistRecord = copyRecord(item)
		shouldPersist = true
	}
	r.mu.Unlock()
	if shouldPersist {
		r.persist(persistRecord)
	}
}

func (r *registryStore) UpdateRisk(subID uint16, uniqueKey uint64, risk uint8) {
	var persistRecord nodeModel.Record
	var shouldPersist bool
	r.mu.Lock()
	key := registryKey{subID: subID, uniqueKey: uniqueKey}
	item, ok := r.items[key]
	if ok && item.Info != nil {
		item.Info.Risk = risk
		item.UpdatedAt = time.Now()
		r.items[key] = item
		persistRecord = copyRecord(item)
		shouldPersist = true
	}
	r.mu.Unlock()
	if shouldPersist {
		r.persist(persistRecord)
	}
}

func (r *registryStore) UpdateSpeed(subID uint16, uniqueKey uint64, up uint32, down uint32, source string) {
	var persistRecord nodeModel.Record
	var shouldPersist bool
	r.mu.Lock()
	key := registryKey{subID: subID, uniqueKey: uniqueKey}
	item, ok := r.items[key]
	if ok && item.Info != nil {
		if up > 0 {
			item.Info.SpeedUp.Update(up)
		}
		if down > 0 {
			item.Info.SpeedDown.Update(down)
		}
		item.LastCheckAt = time.Now()
		item.LastCheckSource = source
		item.UpdatedAt = time.Now()
		r.items[key] = item
		persistRecord = copyRecord(item)
		shouldPersist = true
	}
	r.mu.Unlock()
	if shouldPersist {
		r.persist(persistRecord)
	}
}

func (r *registryStore) UpdateInitStatus(subID uint16, uniqueKey uint64, status nodeModel.InitStatus, reason string) {
	var persistRecord nodeModel.Record
	var shouldPersist bool
	r.mu.Lock()
	key := registryKey{subID: subID, uniqueKey: uniqueKey}
	item, ok := r.items[key]
	if ok {
		item.InitStatus = status
		if reason != "" {
			item.LastFailReason = reason
		}
		if status == nodeModel.InitFailed {
			item.LastCheckAt = time.Now()
			if item.LastCheckSource == "" {
				item.LastCheckSource = "initial"
			}
		}
		item.UpdatedAt = time.Now()
		r.items[key] = item
		persistRecord = copyRecord(item)
		shouldPersist = true
	}
	r.mu.Unlock()
	if shouldPersist {
		r.persist(persistRecord)
	}
}

func (r *registryStore) GetAll() []nodeModel.Record {
	r.mu.RLock()
	result := make([]nodeModel.Record, 0, len(r.items))
	for _, item := range r.items {
		result = append(result, item)
	}
	r.mu.RUnlock()
	return result
}

func (r *registryStore) GetBySubIDs(subIDs []uint16) []nodeModel.Record {
	r.mu.RLock()
	result := make([]nodeModel.Record, 0)
	for key, item := range r.items {
		for _, id := range subIDs {
			if key.subID == id {
				result = append(result, item)
				break
			}
		}
	}
	r.mu.RUnlock()
	return result
}

func (r *registryStore) DeleteBySubID(subID uint16) {
	r.mu.Lock()
	for key := range r.items {
		if key.subID == subID {
			delete(r.items, key)
		}
	}
	r.mu.Unlock()
}

func copyRecord(record nodeModel.Record) nodeModel.Record {
	copied := record
	copied.Base.Raw = append([]byte(nil), record.Base.Raw...)
	if record.Info != nil {
		infoCopy := *record.Info
		infoCopy.Delay = cloneQueue(infoCopy.Delay)
		infoCopy.SpeedUp = cloneQueue(infoCopy.SpeedUp)
		infoCopy.SpeedDown = cloneQueue(infoCopy.SpeedDown)
		copied.Info = &infoCopy
	}
	return copied
}
