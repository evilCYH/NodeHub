package node

import (
	"time"

	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
)

func UpdateRegistryAlive(subID uint16, uniqueKey uint64, alive bool, delay uint16, source string) {
	registry.UpdateAlive(subID, uniqueKey, alive, delay, source)
}

func UpdateRegistryCountry(subID uint16, uniqueKey uint64, country string, ok bool, source string) {
	registry.UpdateCountry(subID, uniqueKey, country, ok, source)
}

func UpdateRegistrySpeed(subID uint16, uniqueKey uint64, up uint32, down uint32, source string) {
	registry.UpdateSpeed(subID, uniqueKey, up, down, source)
}

func UpdateRegistryTikTok(subID uint16, uniqueKey uint64, aliveStatus uint64, source string) {
	record, ok := registry.Get(subID, uniqueKey)
	if !ok || record.Info == nil {
		return
	}
	record.Info.AliveStatus = aliveStatus
	record.LastCheckAt = time.Now()
	record.LastCheckSource = source
	registry.Upsert(record)
}

func UpdateRegistryInitStatus(subID uint16, uniqueKey uint64, status nodeModel.InitStatus, reason string) {
	registry.UpdateInitStatus(subID, uniqueKey, status, reason)
}
