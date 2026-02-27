package op

import (
	"sync"

	"github.com/evilCYH/NodeHub/internal/database/interfaces"
	"github.com/evilCYH/NodeHub/internal/utils/generic"
)

var repo interfaces.Repository

func SetRepo(repository interfaces.Repository) {
	repo = repository
	authRepo = nil
	authData = nil
	checkRepo = nil
	checkCache.Clear()
	nr = nil
	ntr = nil
	notifyCache.Clear()
	notifyTemplateCache.Clear()
	settingRepo = nil
	settingCache.Clear()
	storageRepo = nil
	storageCache.Clear()
	subRepo = nil
	subCache.Clear()
	shareRepo = nil
	shareCache.Clear()
	updateDataBuffer = nil
	pendingUpdates = &generic.MapOf[uint16, bool]{}
	startOnce = sync.Once{}

	shareRepo = repository.Share()
}

func HasRepo() bool {
	return repo != nil
}

func SubRunRepo() interfaces.SubRunRepository {
	return repo.SubRun()
}

func NodeLogRepo() interfaces.NodeLogRepository {
	return repo.NodeLog()
}

func NodeUpdateLogRepo() interfaces.NodeUpdateLogRepository {
	return repo.NodeUpdateLog()
}

func NodeRegistryRepo() interfaces.NodeRegistryRepository {
	return repo.NodeRegistry()
}
func Close() error {
	if repo == nil {
		return nil
	}

	updateAccessCount()
	err := repo.Close()

	repo = nil
	authRepo = nil
	authData = nil
	checkRepo = nil
	checkCache.Clear()
	nr = nil
	ntr = nil
	notifyCache.Clear()
	notifyTemplateCache.Clear()
	settingRepo = nil
	settingCache.Clear()
	storageRepo = nil
	storageCache.Clear()
	subRepo = nil
	subCache.Clear()
	shareRepo = nil
	shareCache.Clear()
	updateDataBuffer = nil
	pendingUpdates = &generic.MapOf[uint16, bool]{}
	startOnce = sync.Once{}

	return err
}
