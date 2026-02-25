package migration

import "github.com/evilCYH/NodeHub/internal/database/migration"

// Migration006AddSubInfoFields 为 sub 表增加流量和到期信息字段
func Migration006AddSubInfoFields() string {
	return `
ALTER TABLE "sub" ADD COLUMN upload INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "sub" ADD COLUMN download INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "sub" ADD COLUMN total INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "sub" ADD COLUMN expire INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "sub" ADD COLUMN info_updated_at DATETIME;
`
}

func init() {
	migration.Register(ClientName, 202602251900, "dev", "Add Sub Info Fields", Migration006AddSubInfoFields)
}
