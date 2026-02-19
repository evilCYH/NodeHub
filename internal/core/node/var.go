package node

import (
	"sync"

	nodeModel "github.com/bestruirui/bestsub/internal/models/node"
)

var (
	poolMutex   sync.RWMutex
	pool        []nodeModel.Data
	nodeExist   *exist
	nodeProcess *exist

	refreshMutex   sync.Mutex
	subInfoMap     = make(map[uint16]nodeModel.SimpleInfo)
	countryInfoMap = make(map[string]nodeModel.SimpleInfo)
	subAggBuf      = make(map[uint16]*infoSums)
	countryAggBuf  = make(map[string]*infoSums)

	failedNodeMu    sync.RWMutex
	failedNodeStore = make(map[uint16][]nodeModel.FailedNode)

	registry = newRegistryStore()
)

type infoSums struct {
	sumSpeedUp   uint64
	sumSpeedDown uint64
	sumDelay     uint64
	sumRisk      uint64
	count        uint32
}
