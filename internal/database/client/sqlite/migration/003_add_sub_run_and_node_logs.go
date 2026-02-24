package migration

import "github.com/evilCYH/NodeHub/internal/database/migration"

// Migration003AddSubRunAndNodeLogs 添加订阅运行日志和节点日志表
func Migration003AddSubRunAndNodeLogs() string {
	return `
CREATE TABLE IF NOT EXISTS "sub_run" (
	"id" INTEGER NOT NULL,
	"sub_id" INTEGER NOT NULL,
	"status" TEXT NOT NULL,
	"message" TEXT,
	"raw_count" INTEGER DEFAULT 0,
	"accepted" INTEGER DEFAULT 0,
	"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	"duration_ms" INTEGER DEFAULT 0,
	PRIMARY KEY("id")
);

CREATE INDEX IF NOT EXISTS idx_sub_run_sub_time ON sub_run(sub_id, created_at);

CREATE TABLE IF NOT EXISTS "sub_run_event" (
	"id" INTEGER NOT NULL,
	"run_id" INTEGER NOT NULL,
	"sub_id" INTEGER NOT NULL,
	"step" TEXT NOT NULL,
	"level" TEXT NOT NULL,
	"message" TEXT NOT NULL,
	"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY("id")
);

CREATE INDEX IF NOT EXISTS idx_sub_run_event_run_time ON sub_run_event(run_id, created_at);
CREATE INDEX IF NOT EXISTS idx_sub_run_event_sub_time ON sub_run_event(sub_id, created_at);

CREATE TABLE IF NOT EXISTS "node_log" (
	"id" INTEGER NOT NULL,
	"sub_id" INTEGER NOT NULL,
	"node_key" INTEGER NOT NULL,
	"node_name" TEXT NOT NULL,
	"level" TEXT NOT NULL,
	"source" TEXT NOT NULL,
	"run_id" INTEGER DEFAULT 0,
	"check_id" INTEGER DEFAULT 0,
	"message" TEXT NOT NULL,
	"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY("id")
);

CREATE INDEX IF NOT EXISTS idx_node_log_sub_time ON node_log(sub_id, created_at);
CREATE INDEX IF NOT EXISTS idx_node_log_sub_source_time ON node_log(sub_id, source, created_at);
CREATE INDEX IF NOT EXISTS idx_node_log_check_time ON node_log(check_id, created_at);

INSERT OR IGNORE INTO setting (key, value) VALUES ('node_log_keep_days', '14');
INSERT OR IGNORE INTO setting (key, value) VALUES ('sub_log_keep_days', '30');
`
}

// init 自动注册迁移
func init() {
	migration.Register(ClientName, 202512161200, "dev", "Add Sub Run and Node Logs", Migration003AddSubRunAndNodeLogs)
}
