package op

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/evilCYH/NodeHub/internal/database/interfaces"
	"github.com/evilCYH/NodeHub/internal/models/setting"
	subModel "github.com/evilCYH/NodeHub/internal/models/sub"
	"github.com/evilCYH/NodeHub/internal/utils/cache"
)

var subRepo interfaces.SubRepository
var subCache = cache.New[uint16, subModel.Data](16)

var ErrSubNotFound = errors.New("sub not found")

type subOrderValidationError struct {
	msg string
}

func (e *subOrderValidationError) Error() string {
	return e.msg
}

func newSubOrderValidationError(msg string) error {
	return &subOrderValidationError{msg: msg}
}

func IsSubOrderValidationError(err error) bool {
	var target *subOrderValidationError
	return errors.As(err, &target)
}

func SubRepo() interfaces.SubRepository {
	if subRepo == nil {
		subRepo = repo.Sub()
	}
	return subRepo
}
func GetSubList(ctx context.Context) ([]subModel.Data, error) {
	subList := subCache.GetAll()
	if len(subList) == 0 {
		err := refreshSubCache(context.Background())
		if err != nil {
			return nil, err
		}
		subList = subCache.GetAll()
	}
	var result = make([]subModel.Data, 0, len(subList))
	for _, v := range subList {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].SortOrder == result[j].SortOrder {
			return result[i].ID < result[j].ID
		}
		return result[i].SortOrder < result[j].SortOrder
	})
	return result, nil
}

func GetSubByID(ctx context.Context, id uint16) (*subModel.Data, error) {
	if subCache.Len() == 0 {
		if err := refreshSubCache(ctx); err != nil {
			return nil, err
		}
	}
	if s, ok := subCache.Get(id); ok {
		return &s, nil
	}
	return nil, ErrSubNotFound
}
func GetSubNameByID(ctx context.Context, id uint16) string {
	sub, err := GetSubByID(ctx, id)
	if err != nil {
		return ""
	}
	return sub.Name
}
func GetSubTagsByID(ctx context.Context, id uint16) []string {
	sub, err := GetSubByID(ctx, id)
	if err != nil {
		return []string{}
	}
	tags := make([]string, 0)
	err = json.Unmarshal([]byte(sub.Tags), &tags)
	if err != nil {
		return []string{}
	}
	return tags
}
func CreateSub(ctx context.Context, sub *subModel.Data) error {
	if subCache.Len() == 0 {
		if err := refreshSubCache(ctx); err != nil {
			return err
		}
	}
	if err := SubRepo().Create(ctx, sub); err != nil {
		return err
	}
	subCache.Set(sub.ID, *sub)
	return nil
}

func BatchCreateSub(ctx context.Context, subs []*subModel.Data) error {
	if subCache.Len() == 0 {
		if err := refreshSubCache(ctx); err != nil {
			return err
		}
	}
	if err := SubRepo().BatchCreate(ctx, subs); err != nil {
		return err
	}
	for _, sub := range subs {
		subCache.Set(sub.ID, *sub)
	}
	return nil
}
func UpdateSub(ctx context.Context, sub *subModel.Data) error {
	if subCache.Len() == 0 {
		if err := refreshSubCache(ctx); err != nil {
			return err
		}
	}
	oldSub, ok := subCache.Get(sub.ID)
	if !ok {
		return ErrSubNotFound
	}
	sub.Result = oldSub.Result
	sub.SortOrder = oldSub.SortOrder
	sub.CreatedAt = oldSub.CreatedAt
	sub.Upload = oldSub.Upload
	sub.Download = oldSub.Download
	sub.Total = oldSub.Total
	sub.Expire = oldSub.Expire
	sub.InfoUpdatedAt = oldSub.InfoUpdatedAt
	if err := SubRepo().Update(ctx, sub); err != nil {
		return err
	}
	subCache.Set(sub.ID, *sub)
	return nil
}
func UpdateSubResult(ctx context.Context, id uint16, result subModel.Result) error {
	if subCache.Len() == 0 {
		if err := refreshSubCache(ctx); err != nil {
			return err
		}
	}
	sub, ok := subCache.Get(id)
	if !ok {
		return ErrSubNotFound
	}
	var oldStatus subModel.Result
	json.Unmarshal([]byte(sub.Result), &oldStatus)

	result.LastStatus = "success"
	if result.Fail > 0 {
		result.LastStatus = "error"
	}
	result.Success += oldStatus.Success
	result.Fail += oldStatus.Fail
	if result.NodeNullCount != 0 {
		result.NodeNullCount += oldStatus.NodeNullCount
	}
	subDisableThreshold := GetSettingInt(setting.SUB_DISABLE_AUTO)
	if subDisableThreshold > 0 && result.NodeNullCount > uint32(subDisableThreshold) {
		sub.Enable = false
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		return err
	}
	sub.Result = string(bytes)
	if err := SubRepo().Update(ctx, &sub); err != nil {
		return err
	}
	subCache.Set(id, sub)
	return nil
}
func DeleteSub(ctx context.Context, id uint16) error {
	if subCache.Len() == 0 {
		if err := refreshSubCache(ctx); err != nil {
			return err
		}
	}
	deleted, err := SubRepo().DeleteCascade(ctx, id)
	if err != nil {
		return err
	}
	subCache.Del(id)
	if !deleted {
		return ErrSubNotFound
	}
	return nil
}

func UpdateSubInfo(ctx context.Context, id uint16, upload, download, total, expire int64, infoUpdatedAt *time.Time) error {
	if subCache.Len() == 0 {
		if err := refreshSubCache(ctx); err != nil {
			return err
		}
	}
	sub, ok := subCache.Get(id)
	if !ok {
		return ErrSubNotFound
	}
	if err := SubRepo().UpdateSubInfo(ctx, id, upload, download, total, expire, infoUpdatedAt); err != nil {
		return err
	}
	sub.Upload = upload
	sub.Download = download
	sub.Total = total
	sub.Expire = expire
	sub.InfoUpdatedAt = infoUpdatedAt
	subCache.Set(id, sub)
	return nil
}

func UpdateSubOrder(ctx context.Context, orders []subModel.SortOrderItem) error {
	if len(orders) == 0 {
		return newSubOrderValidationError("orders is empty")
	}
	if subCache.Len() == 0 {
		if err := refreshSubCache(ctx); err != nil {
			return err
		}
	}

	subMap := subCache.GetAll()
	if len(orders) != len(subMap) {
		return newSubOrderValidationError("orders must include all subscriptions")
	}

	seenIDs := make(map[uint16]struct{}, len(orders))
	seenSortOrders := make([]bool, len(orders))
	for _, item := range orders {
		if _, ok := subMap[item.ID]; !ok {
			return newSubOrderValidationError(fmt.Sprintf("subscription not found: %d", item.ID))
		}
		if _, duplicated := seenIDs[item.ID]; duplicated {
			return newSubOrderValidationError(fmt.Sprintf("duplicate subscription id: %d", item.ID))
		}
		seenIDs[item.ID] = struct{}{}

		if item.SortOrder < 0 || item.SortOrder >= len(orders) {
			return newSubOrderValidationError(fmt.Sprintf("invalid sort_order: %d", item.SortOrder))
		}
		if seenSortOrders[item.SortOrder] {
			return newSubOrderValidationError(fmt.Sprintf("duplicate sort_order: %d", item.SortOrder))
		}
		seenSortOrders[item.SortOrder] = true
	}
	for index, exists := range seenSortOrders {
		if !exists {
			return newSubOrderValidationError(fmt.Sprintf("missing sort_order: %d", index))
		}
	}

	if err := SubRepo().UpdateSortOrder(ctx, orders); err != nil {
		return err
	}
	for _, item := range orders {
		cachedSub := subMap[item.ID]
		cachedSub.SortOrder = item.SortOrder
		subCache.Set(item.ID, cachedSub)
	}
	return nil
}

func refreshSubCache(ctx context.Context) error {
	subList, err := SubRepo().List(ctx)
	if err != nil {
		return err
	}
	for _, s := range *subList {
		subCache.Set(s.ID, s)
	}
	return nil
}
