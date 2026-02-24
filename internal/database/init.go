package database

import (
	"context"

	"github.com/evilCYH/NodeHub/internal/database/interfaces"
	"github.com/evilCYH/NodeHub/internal/database/op"
	authModel "github.com/evilCYH/NodeHub/internal/models/auth"
	"github.com/evilCYH/NodeHub/internal/models/notify"
	"github.com/evilCYH/NodeHub/internal/models/setting"
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"golang.org/x/crypto/bcrypt"
)

func initAuth(ctx context.Context, auth interfaces.AuthRepository) error {
	isInitialized, err := auth.IsInitialized(ctx)
	if err != nil {
		log.Fatalf("failed to check if database is initialized: %v", err)
	}
	if !isInitialized {
		authData := authModel.Default()
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(authData.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("failed to hash password: %v", err)
		}
		authData.Password = string(hashedBytes)
		if err := auth.Initialize(ctx, &authData); err != nil {
			log.Fatalf("failed to initialize auth: %v", err)
		}
		log.Info("初始化默认管理员账号 用户名: admin 密码: admin")
	}
	return nil
}
func initSystemSetting(ctx context.Context, systemSetting interfaces.SettingRepository) error {
	defaultSystemSetting := setting.DefaultSetting()
	existingSystemSetting, err := op.GetAllSetting(ctx)
	notExistSetting := make([]setting.Setting, 0)
	if err != nil {
		log.Fatalf("failed to get existing system setting: %v", err)
	}

	existingSystemSettingMap := make(map[string]bool)
	updateSetting := make([]setting.Setting, 0)
	for _, item := range existingSystemSetting {
		existingSystemSettingMap[item.Key] = true
	}
	if len(updateSetting) > 0 {
		if err := systemSetting.Update(ctx, &updateSetting); err != nil {
			log.Fatalf("failed to update system setting: %v", err)
		}
	}

	for _, s := range defaultSystemSetting {
		if !existingSystemSettingMap[s.Key] {
			notExistSetting = append(notExistSetting, s)
		}
	}

	if len(notExistSetting) > 0 {
		if err := systemSetting.Create(ctx, &notExistSetting); err != nil {
			log.Fatalf("failed to create missing system setting: %v", err)
		}
	}
	return nil
}
func initNotifyTemplate(ctx context.Context, notifyTemplateRepo interfaces.NotifyTemplateRepository) error {
	defaultNotifyTemplates := notify.DefaultTemplates()
	existingNotifyTemplates, err := notifyTemplateRepo.List(ctx)
	if err != nil {
		log.Fatalf("failed to get existing notify templates: %v", err)
	}
	existingNotifyTemplatesMap := make(map[string]bool)
	for _, template := range *existingNotifyTemplates {
		existingNotifyTemplatesMap[template.Type] = true
	}
	for _, template := range defaultNotifyTemplates {
		if !existingNotifyTemplatesMap[template.Type] {
			if err := notifyTemplateRepo.Create(ctx, &template); err != nil {
				log.Fatalf("failed to create missing notify template %s: %v", template.Type, err)
			}
		}
	}
	return nil
}
