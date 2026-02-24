package migration

import "github.com/evilCYH/NodeHub/internal/database/migration"

// Migration004AddNodeUpdateLog 添加节点聚合统计日志表
func Migration004AddNodeUpdateLog() string {
	return `
CREATE TABLE IF NOT EXISTS "node_update_log" (
	"id" INTEGER NOT NULL,
	"sub_id" INTEGER NOT NULL,
	"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	"duration_ms" INTEGER DEFAULT 0,
	"raw_count" INTEGER DEFAULT 0,
	"candidate" INTEGER DEFAULT 0,
	"duplicate" INTEGER DEFAULT 0,
	"invalid" INTEGER DEFAULT 0,
	"test_failed" INTEGER DEFAULT 0,
	"accepted" INTEGER DEFAULT 0,
	"merged" INTEGER DEFAULT 0,
	"dropped" INTEGER DEFAULT 0,
	"details" TEXT,
	PRIMARY KEY("id")
);

CREATE INDEX IF NOT EXISTS idx_node_update_log_sub_time ON node_update_log(sub_id, created_at);
`
}

func init() {
	migration.Register(ClientName, 202512170900, "dev", "Add Node Update Logs", Migration004AddNodeUpdateLog)
}
