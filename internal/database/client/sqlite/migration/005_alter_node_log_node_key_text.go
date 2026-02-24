package migration

import "github.com/evilCYH/NodeHub/internal/database/migration"

// Migration005AlterNodeLogNodeKeyText 迁移 node_log.node_key 为 TEXT
func Migration005AlterNodeLogNodeKeyText() string {
	return `
ALTER TABLE node_log RENAME TO node_log_old;

CREATE TABLE IF NOT EXISTS "node_log" (
	"id" INTEGER NOT NULL,
	"sub_id" INTEGER NOT NULL,
	"node_key" TEXT NOT NULL,
	"node_name" TEXT NOT NULL,
	"level" TEXT NOT NULL,
	"source" TEXT NOT NULL,
	"run_id" INTEGER DEFAULT 0,
	"check_id" INTEGER DEFAULT 0,
	"message" TEXT NOT NULL,
	"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY("id")
);

INSERT INTO node_log (id, sub_id, node_key, node_name, level, source, run_id, check_id, message, created_at)
SELECT id, sub_id, CAST(node_key AS TEXT), node_name, level, source, run_id, check_id, message, created_at
FROM node_log_old;

DROP TABLE node_log_old;

CREATE INDEX IF NOT EXISTS idx_node_log_sub_time ON node_log(sub_id, created_at);
CREATE INDEX IF NOT EXISTS idx_node_log_sub_source_time ON node_log(sub_id, source, created_at);
CREATE INDEX IF NOT EXISTS idx_node_log_check_time ON node_log(check_id, created_at);
`
}

func init() {
	migration.Register(ClientName, 202512171000, "dev", "Alter Node Log Node Key To Text", Migration005AlterNodeLogNodeKeyText)
}
