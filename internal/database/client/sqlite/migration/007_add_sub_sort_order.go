package migration

import "github.com/evilCYH/NodeHub/internal/database/migration"

// Migration007AddSubSortOrder 为 sub 表增加排序字段并初始化历史数据
func Migration007AddSubSortOrder() string {
	return `
ALTER TABLE "sub" ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;
UPDATE "sub" SET sort_order = id WHERE sort_order = 0;
CREATE INDEX IF NOT EXISTS "idx_sub_sort_order" ON "sub" ("sort_order", "id");
`
}

func init() {
	migration.Register(ClientName, 202602261300, "dev", "Add Sub Sort Order", Migration007AddSubSortOrder)
}
