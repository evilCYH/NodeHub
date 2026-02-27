package migration

import "github.com/evilCYH/NodeHub/internal/database/migration"

// Migration008AddNodeRegistry 添加节点主表持久化结构
func Migration008AddNodeRegistry() string {
	return `
CREATE TABLE IF NOT EXISTS "node_registry" (
	"sub_id" INTEGER NOT NULL,
	"unique_key" TEXT NOT NULL,
	"raw" BLOB NOT NULL,
	"info" TEXT,
	"init_status" TEXT NOT NULL,
	"last_check_at" DATETIME,
	"last_check_source" TEXT,
	"last_fail_reason" TEXT,
	"first_seen_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	"updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY("sub_id", "unique_key")
);

CREATE INDEX IF NOT EXISTS idx_node_registry_sub_updated ON node_registry(sub_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_node_registry_last_check ON node_registry(last_check_at);
`
}

func init() {
	migration.Register(ClientName, 202602261500, "dev", "Add Node Registry", Migration008AddNodeRegistry)
}
