package node

import (
	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
	"github.com/evilCYH/NodeHub/internal/utils/generic"
)

func snapshotFromInfo(info *nodeModel.Info) *nodeModel.RegistryInfoSnapshot {
	if info == nil {
		return nil
	}
	return &nodeModel.RegistryInfoSnapshot{
		DelaySamples:     append([]uint16(nil), info.Delay.GetAll()...),
		SpeedUpSamples:   append([]uint32(nil), info.SpeedUp.GetAll()...),
		SpeedDownSamples: append([]uint32(nil), info.SpeedDown.GetAll()...),
		Risk:             info.Risk,
		AliveStatus:      info.AliveStatus,
		IP:               info.IP,
		Country:          info.Country,
	}
}

func infoFromSnapshot(snapshot *nodeModel.RegistryInfoSnapshot) *nodeModel.Info {
	if snapshot == nil {
		return nil
	}
	info := &nodeModel.Info{
		Delay:       buildUint16Queue(snapshot.DelaySamples),
		SpeedUp:     buildUint32Queue(snapshot.SpeedUpSamples),
		SpeedDown:   buildUint32Queue(snapshot.SpeedDownSamples),
		Risk:        snapshot.Risk,
		AliveStatus: snapshot.AliveStatus,
		IP:          snapshot.IP,
		Country:     snapshot.Country,
	}
	return info
}

func buildUint16Queue(samples []uint16) generic.Queue[uint16] {
	capacity := 5
	if len(samples) > capacity {
		capacity = len(samples)
	}
	queue := generic.NewQueue[uint16](capacity)
	for _, value := range samples {
		queue.Update(value)
	}
	return *queue
}

func buildUint32Queue(samples []uint32) generic.Queue[uint32] {
	capacity := 5
	if len(samples) > capacity {
		capacity = len(samples)
	}
	queue := generic.NewQueue[uint32](capacity)
	for _, value := range samples {
		queue.Update(value)
	}
	return *queue
}

func toDBRecord(record nodeModel.Record) nodeModel.RegistryRecordDB {
	raw := append([]byte(nil), record.Base.Raw...)
	return nodeModel.RegistryRecordDB{
		SubID:           record.Base.SubId,
		UniqueKey:       record.Base.UniqueKey,
		Raw:             raw,
		Info:            snapshotFromInfo(record.Info),
		InitStatus:      record.InitStatus,
		LastCheckAt:     record.LastCheckAt,
		LastCheckSource: record.LastCheckSource,
		LastFailReason:  record.LastFailReason,
		FirstSeenAt:     record.FirstSeenAt,
		UpdatedAt:       record.UpdatedAt,
	}
}

func fromDBRecord(record nodeModel.RegistryRecordDB) nodeModel.Record {
	raw := append([]byte(nil), record.Raw...)
	initStatus := record.InitStatus
	if initStatus == "" {
		initStatus = nodeModel.InitUnknown
	}
	return nodeModel.Record{
		Base: nodeModel.Base{
			SubId:     record.SubID,
			UniqueKey: record.UniqueKey,
			Raw:       raw,
		},
		Info:            infoFromSnapshot(record.Info),
		InitStatus:      initStatus,
		LastCheckAt:     record.LastCheckAt,
		LastCheckSource: record.LastCheckSource,
		LastFailReason:  record.LastFailReason,
		FirstSeenAt:     record.FirstSeenAt,
		UpdatedAt:       record.UpdatedAt,
	}
}
