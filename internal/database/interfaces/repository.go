package interfaces

// Repository 统一的仓库接口
type Repository interface {
	Auth() AuthRepository

	Setting() SettingRepository

	Notify() NotifyRepository
	NotifyTemplate() NotifyTemplateRepository

	Check() CheckRepository

	Sub() SubRepository
	Share() ShareRepository
	SubRun() SubRunRepository
	NodeLog() NodeLogRepository
	NodeUpdateLog() NodeUpdateLogRepository
	NodeRegistry() NodeRegistryRepository

	Storage() StorageRepository

	Close() error
	Migrate() error
}
