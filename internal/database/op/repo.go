package op

import (
	"github.com/bestruirui/bestsub/internal/database/interfaces"
)

var repo interfaces.Repository

func SetRepo(repository interfaces.Repository) {
	repo = repository
	shareRepo = repository.Share()
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
func Close() error {
	updateAccessCount()
	return repo.Close()
}
