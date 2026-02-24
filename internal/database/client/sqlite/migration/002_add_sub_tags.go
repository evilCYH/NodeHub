package migration

import "github.com/evilCYH/NodeHub/internal/database/migration"

// Migration002AddSubTags 为sub表添加列tags
func Migration002AddSubTags() string {
	return `
ALTER TABLE "sub"
ADD COLUMN tags TEXT NOT NULL DEFAULT '[]';
`
}

// init 自动注册迁移
func init() {
	migration.Register(ClientName, 202511082145, "dev", "Add Sub Tags", Migration002AddSubTags)
}
