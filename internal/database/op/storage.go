package op

import (
	"context"
	"fmt"

	"github.com/evilCYH/NodeHub/internal/database/interfaces"
	"github.com/evilCYH/NodeHub/internal/models/storage"
	"github.com/evilCYH/NodeHub/internal/utils/cache"
)

var storageRepo interfaces.StorageRepository
var storageCache = cache.New[uint16, storage.Data](16)

func StorageRepo() interfaces.StorageRepository {
	if storageRepo == nil {
		storageRepo = repo.Storage()
	}
	return storageRepo
}
func GetStorageList(ctx context.Context) ([]storage.Data, error) {
	storageList := storageCache.GetAll()
	if len(storageList) == 0 {
		err := refreshStorageCache(context.Background())
		if err != nil {
			return nil, err
		}
		storageList = storageCache.GetAll()
	}
	var result = make([]storage.Data, 0, len(storageList))
	for _, v := range storageList {
		result = append(result, v)
	}
	return result, nil
}

func GetStorageByID(ctx context.Context, id uint16) (*storage.Data, error) {
	if storageCache.Len() == 0 {
		if err := refreshStorageCache(ctx); err != nil {
			return nil, err
		}
	}
	if s, ok := storageCache.Get(id); ok {
		return &s, nil
	}
	return nil, fmt.Errorf("storage not found")
}
func CreateStorage(ctx context.Context, storage *storage.Data) error {
	if storageCache.Len() == 0 {
		if err := refreshStorageCache(ctx); err != nil {
			return err
		}
	}
	if err := StorageRepo().Create(ctx, storage); err != nil {
		return err
	}
	storageCache.Set(storage.ID, *storage)
	return nil
}
func UpdateStorage(ctx context.Context, storage *storage.Data) error {
	if storageCache.Len() == 0 {
		if err := refreshStorageCache(ctx); err != nil {
			return err
		}
	}
	if err := StorageRepo().Update(ctx, storage); err != nil {
		return err
	}
	storageCache.Set(storage.ID, *storage)
	return nil
}

func DeleteStorage(ctx context.Context, id uint16) error {
	if storageCache.Len() == 0 {
		if err := refreshStorageCache(ctx); err != nil {
			return err
		}
	}
	if err := StorageRepo().Delete(ctx, id); err != nil {
		return err
	}
	storageCache.Del(id)
	return nil
}

func refreshStorageCache(ctx context.Context) error {
	storageList, err := StorageRepo().List(ctx)
	if err != nil {
		return err
	}
	for _, s := range *storageList {
		storageCache.Set(s.ID, s)
	}
	return nil
}
