package op

import (
	"context"

	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
)

func UpsertNodeRegistry(ctx context.Context, record *nodeModel.RegistryRecordDB) error {
	return repo.NodeRegistry().Upsert(ctx, record)
}

func ListNodeRegistry(ctx context.Context) ([]nodeModel.RegistryRecordDB, error) {
	return repo.NodeRegistry().List(ctx)
}

func DeleteNodeRegistryBySubID(ctx context.Context, subID uint16) error {
	return repo.NodeRegistry().DeleteBySubID(ctx, subID)
}
