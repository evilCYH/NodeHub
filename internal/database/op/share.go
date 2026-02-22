package op

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bestruirui/bestsub/internal/database/interfaces"
	"github.com/bestruirui/bestsub/internal/models/share"
	"github.com/bestruirui/bestsub/internal/utils/cache"
	"github.com/bestruirui/bestsub/internal/utils/generic"
	"github.com/bestruirui/bestsub/internal/utils/log"
)

var shareRepo interfaces.ShareRepository
var shareCache = cache.New[uint16, share.Data](16)

var pendingUpdates = &generic.MapOf[uint16, bool]{}
var startOnce sync.Once
var updateMu sync.Mutex

func ShareRepo() interfaces.ShareRepository {
	if shareRepo == nil {
		shareRepo = repo.Share()
	}
	return shareRepo
}
func GetShareList(ctx context.Context) ([]share.Data, error) {
	shareList := shareCache.GetAll()
	if len(shareList) == 0 {
		err := refreshShareCache(context.Background())
		if err != nil {
			return nil, err
		}
		shareList = shareCache.GetAll()
	}
	var result = make([]share.Data, 0, len(shareList))
	for _, v := range shareList {
		result = append(result, v)
	}
	return result, nil
}

func GetShareByID(ctx context.Context, id uint16) (*share.Data, error) {
	if shareCache.Len() == 0 {
		if err := refreshShareCache(ctx); err != nil {
			return nil, err
		}
	}
	if s, ok := shareCache.Get(id); ok {
		return &s, nil
	}
	return nil, fmt.Errorf("share not found")
}
func GetShareByToken(ctx context.Context, token string) (*share.Data, error) {
	if shareCache.Len() == 0 {
		if err := refreshShareCache(ctx); err != nil {
			return nil, err
		}
	}
	for _, s := range shareCache.GetAll() {
		if s.Token == token {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("share not found")
}
func CreateShare(ctx context.Context, share *share.Data) error {
	if shareCache.Len() == 0 {
		if err := refreshShareCache(ctx); err != nil {
			return err
		}
	}
	if err := ShareRepo().Create(ctx, share); err != nil {
		return err
	}
	shareCache.Set(share.ID, *share)
	return nil
}
func UpdateShare(ctx context.Context, share *share.Data) error {
	if shareCache.Len() == 0 {
		if err := refreshShareCache(ctx); err != nil {
			return err
		}
	}
	oldShare, ok := shareCache.Get(share.ID)
	if !ok {
		return fmt.Errorf("share not found")
	}
	share.AccessCount = oldShare.AccessCount
	if err := ShareRepo().Update(ctx, share); err != nil {
		return err
	}
	shareCache.Set(share.ID, *share)
	return nil
}

func UpdateShareAccessCount(ctx context.Context, id uint16) error {
	if shareCache.Len() == 0 {
		if err := refreshShareCache(ctx); err != nil {
			return err
		}
	}
	share, ok := shareCache.Get(id)
	if !ok {
		return fmt.Errorf("share not found")
	}
	share.AccessCount++
	shareCache.Set(id, share)

	pendingUpdates.Store(id, true)

	if shareRepo != nil {
		if err := ShareRepo().UpdateAccessCount(ctx, &[]share.UpdateAccessCountDB{
			{ID: id, AccessCount: share.AccessCount},
		}); err != nil {
			log.Errorf("failed to update share access count: %v", err)
			startScheduleUpdateAccessCount()
		} else {
			pendingUpdates.Delete(id)
		}
	} else {
		startScheduleUpdateAccessCount()
	}

	return nil
}

func DeleteShare(ctx context.Context, id uint16) error {
	if shareCache.Len() == 0 {
		if err := refreshShareCache(ctx); err != nil {
			return err
		}
	}
	if err := ShareRepo().Delete(ctx, id); err != nil {
		return err
	}
	shareCache.Del(id)
	return nil
}

func refreshShareCache(ctx context.Context) error {
	shareList, err := ShareRepo().List(ctx)
	if err != nil {
		return err
	}
	for _, s := range *shareList {
		shareCache.Set(s.ID, s)
	}
	return nil
}

func startScheduleUpdateAccessCount() {
	startOnce.Do(func() {
		ticker := time.NewTicker(60 * time.Second)
		go func() {
			defer ticker.Stop()
			for range ticker.C {
				updateAccessCount()
			}
		}()
	})
}

var updateDataBuffer []share.UpdateAccessCountDB

func updateAccessCount() {
	updateMu.Lock()
	defer updateMu.Unlock()

	updateDataBuffer = updateDataBuffer[:0]

	pendingUpdates.Range(func(id uint16, _ bool) bool {
		if shareData, ok := shareCache.Get(id); ok {
			updateDataBuffer = append(updateDataBuffer, share.UpdateAccessCountDB{
				ID:          id,
				AccessCount: shareData.AccessCount,
			})
		}
		return true
	})
	if len(updateDataBuffer) == 0 {
		return
	}
	if err := ShareRepo().UpdateAccessCount(context.Background(), &updateDataBuffer); err != nil {
		log.Errorf("failed to update share access count: %v", err)
		return
	}
	for _, data := range updateDataBuffer {
		pendingUpdates.Delete(data.ID)
	}
}
