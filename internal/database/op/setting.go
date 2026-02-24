package op

import (
	"context"
	"fmt"
	"strconv"

	"github.com/evilCYH/NodeHub/internal/database/interfaces"
	"github.com/evilCYH/NodeHub/internal/models/setting"
	"github.com/evilCYH/NodeHub/internal/utils/cache"
)

var settingRepo interfaces.SettingRepository
var settingCache = cache.New[string, string](4)

func SettingRepo() interfaces.SettingRepository {
	if settingRepo == nil {
		settingRepo = repo.Setting()
	}
	return settingRepo
}
func GetAllSettingMap(ctx context.Context) (map[string]string, error) {
	sysConfCache := settingCache.GetAll()
	if len(sysConfCache) == 0 {
		err := refreshSettingCache(context.Background())
		if err != nil {
			return nil, err
		}
		sysConfCache = settingCache.GetAll()
	}
	return sysConfCache, nil
}
func GetAllSetting(ctx context.Context) ([]setting.Setting, error) {
	sysConfCache := settingCache.GetAll()
	if len(sysConfCache) == 0 {
		err := refreshSettingCache(context.Background())
		if err != nil {
			return nil, err
		}
		sysConfCache = settingCache.GetAll()
	}
	var result []setting.Setting
	for key, value := range sysConfCache {
		result = append(result, setting.Setting{
			Key:   key,
			Value: value,
		})
	}
	return result, nil
}
func GetSettingByKey(key string) (string, error) {
	if value, ok := settingCache.Get(key); ok {
		return value, nil
	}
	err := refreshSettingCache(context.Background())
	if err != nil {
		return "", err
	}
	if value, ok := settingCache.Get(key); ok {
		return value, nil
	}
	return "", fmt.Errorf("config not found")
}
func UpdateSetting(ctx context.Context, setting *[]setting.Setting) error {
	if settingCache.Len() == 0 {
		err := refreshSettingCache(context.Background())
		if err != nil {
			return err
		}
	}
	if err := SettingRepo().Update(ctx, setting); err != nil {
		return err
	}
	for _, item := range *setting {
		settingCache.Set(item.Key, item.Value)
	}
	return nil

}
func GetSettingStr(key string) string {
	value, err := GetSettingByKey(key)
	if err != nil {
		return ""
	}
	return value
}
func GetSettingInt(key string) int {
	value, err := GetSettingByKey(key)
	if err != nil {
		return 0
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return i
}
func GetSettingBool(key string) bool {
	value, err := GetSettingByKey(key)
	if err != nil {
		return false
	}
	return value == "true"
}

func refreshSettingCache(ctx context.Context) error {
	settingCache.Clear()
	configs, err := SettingRepo().GetAll(ctx)
	if err != nil {
		return err
	}
	for _, config := range *configs {
		settingCache.Set(config.Key, config.Value)
	}
	return nil
}
