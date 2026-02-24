package node

import (
	"sync"
	"time"

	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
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
		record.SeenAt = existing.SeenAt
		record.UpdatedAt = now
	} else {
		record.SeenAt = now
		record.UpdatedAt = now
	}
	r.items[key] = record
	r.mu.Unlock()
}

func (r *registryStore) Get(subID uint16, uniqueKey uint64) (nodeModel.Record, bool) {
	r.mu.RLock()
	item, ok := r.items[registryKey{subID: subID, uniqueKey: uniqueKey}]
	r.mu.RUnlock()
	return item, ok
}

func (r *registryStore) UpdateInfo(subID uint16, uniqueKey uint64, info *nodeModel.Info, source string) {
	r.mu.Lock()
	key := registryKey{subID: subID, uniqueKey: uniqueKey}
	item, ok := r.items[key]
	if ok {
		item.Info = info
		item.LastCheckAt = time.Now()
		item.LastCheckSource = source
		item.UpdatedAt = time.Now()
		r.items[key] = item
	}
	r.mu.Unlock()
}

func (r *registryStore) UpdateAlive(subID uint16, uniqueKey uint64, alive bool, delay uint16, source string) {
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
	}
	r.mu.Unlock()
}

func (r *registryStore) UpdateCountry(subID uint16, uniqueKey uint64, country string, ok bool, source string) {
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
	}
	r.mu.Unlock()
}

func (r *registryStore) UpdateRisk(subID uint16, uniqueKey uint64, risk uint8) {
	r.mu.Lock()
	key := registryKey{subID: subID, uniqueKey: uniqueKey}
	item, ok := r.items[key]
	if ok && item.Info != nil {
		item.Info.Risk = risk
		item.UpdatedAt = time.Now()
		r.items[key] = item
	}
	r.mu.Unlock()
}

func (r *registryStore) UpdateSpeed(subID uint16, uniqueKey uint64, up uint32, down uint32, source string) {
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
	}
	r.mu.Unlock()
}

func (r *registryStore) UpdateInitStatus(subID uint16, uniqueKey uint64, status nodeModel.InitStatus, reason string) {
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
	}
	r.mu.Unlock()
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
