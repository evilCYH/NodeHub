package op

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/evilCYH/NodeHub/internal/database/interfaces"
	"github.com/evilCYH/NodeHub/internal/models/check"
	"github.com/evilCYH/NodeHub/internal/utils/cache"
	"github.com/evilCYH/NodeHub/internal/utils/log"
)

var checkRepo interfaces.CheckRepository
var checkCache = cache.New[uint16, check.Data](16)

func CheckRepo() interfaces.CheckRepository {
	if checkRepo == nil {
		checkRepo = repo.Check()
	}
	return checkRepo
}
func GetCheckByID(id uint16) (check.Data, error) {
	if checkCache.Len() == 0 {
		if err := refreshCheckCache(); err != nil {
			return check.Data{}, err
		}
	}
	if t, ok := checkCache.Get(id); ok {
		return t, nil
	}
	return check.Data{}, fmt.Errorf("check not found")
}
func CreateCheck(ctx context.Context, t *check.Data) error {
	if checkCache.Len() == 0 {
		if err := refreshCheckCache(); err != nil {
			return err
		}
	}
	if err := CheckRepo().Create(ctx, t); err != nil {
		return err
	}
	checkCache.Set(t.ID, *t)
	return nil
}
func UpdateCheck(ctx context.Context, t *check.Data) error {
	if checkCache.Len() == 0 {
		if err := refreshCheckCache(); err != nil {
			return err
		}
	}
	oldCheck, ok := checkCache.Get(t.ID)
	if !ok {
		return fmt.Errorf("task not found")
	}
	t.Result = oldCheck.Result
	if err := CheckRepo().Update(ctx, t); err != nil {
		return err
	}
	checkCache.Set(t.ID, *t)
	return nil
}
func UpdateCheckResult(id uint16, result check.Result) error {
	if checkCache.Len() == 0 {
		if err := refreshCheckCache(); err != nil {
			log.Errorf("failed to refresh check cache: %v", err)
			return err
		}
	}
	oldCheck, ok := checkCache.Get(id)
	if !ok {
		log.Errorf("check not found")
		return fmt.Errorf("task not found")
	}
	var oldResult check.Result
	if oldCheck.Result != "" {
		if err := json.Unmarshal([]byte(oldCheck.Result), &oldResult); err != nil {
			log.Errorf("failed to unmarshal check result: %v", err)
			return err
		}
	}
	oldResult.Msg = result.Msg
	oldResult.Extra = result.Extra
	oldResult.LastRun = time.Now()
	oldResult.Duration = result.Duration
	resultBytes, err := json.Marshal(oldResult)
	if err != nil {
		log.Errorf("failed to marshal check result: %v", err)
		return err
	}
	oldCheck.Result = string(resultBytes)
	if err := CheckRepo().Update(context.Background(), &oldCheck); err != nil {
		log.Errorf("failed to update check result: %v", err)
		return err
	}
	checkCache.Set(id, oldCheck)
	return nil
}
func DeleteCheck(ctx context.Context, id uint16) error {
	if checkCache.Len() == 0 {
		if err := refreshCheckCache(); err != nil {
			return err
		}
	}
	if err := CheckRepo().Delete(ctx, id); err != nil {
		return err
	}
	checkCache.Del(id)
	return nil
}
func GetCheckList() ([]check.Data, error) {
	taskList := checkCache.GetAll()
	if len(taskList) == 0 {
		err := refreshCheckCache()
		if err != nil {
			return nil, err
		}
		taskList = checkCache.GetAll()
	}
	var result = make([]check.Data, 0, len(taskList))
	for _, v := range taskList {
		result = append(result, v)
	}
	return result, nil
}
func refreshCheckCache() error {
	checkCache.Clear()
	checks, err := CheckRepo().List(context.Background())
	if err != nil {
		return err
	}
	for _, check := range *checks {
		checkCache.Set(check.ID, check)
	}
	return nil
}
