package migration

import "github.com/evilCYH/NodeHub/internal/database/migration"

// Migration001Table 初始数据库架构
func Migration001Table() string {
	return `
CREATE TABLE IF NOT EXISTS "auth" (
	"id" INTEGER,
	"username" TEXT NOT NULL UNIQUE,
	"password" TEXT NOT NULL,
	PRIMARY KEY("id")
);

CREATE TABLE IF NOT EXISTS "setting" (
	"key" TEXT NOT NULL UNIQUE,
	"value" TEXT NOT NULL,
	PRIMARY KEY("key")
);

CREATE TABLE IF NOT EXISTS "notify_template" (
	"type" TEXT NOT NULL,
	"template" TEXT NOT NULL,
	PRIMARY KEY("type")
);

CREATE TABLE IF NOT EXISTS "notify" (
	"id" INTEGER NOT NULL UNIQUE,
	"name" TEXT NOT NULL,
	"type" TEXT NOT NULL,
	"config" TEXT NOT NULL,
	PRIMARY KEY("id")
);

CREATE TABLE IF NOT EXISTS "check_task" (
	"id" INTEGER,
	"enable" BOOLEAN NOT NULL,
	"name" TEXT,
	"task" TEXT NOT NULL,
	"config" TEXT,
	"result" TEXT,
	PRIMARY KEY("id")
);

CREATE TABLE IF NOT EXISTS "storage" (
	"id" INTEGER,
	"name" TEXT,
	"type" TEXT NOT NULL,
	"config" TEXT NOT NULL,
	PRIMARY KEY("id")
);

CREATE TABLE IF NOT EXISTS "sub_template" (
	"id" INTEGER,
	"name" TEXT,
	"type" TEXT NOT NULL,
	"template" TEXT NOT NULL,
	PRIMARY KEY("id")
);

CREATE TABLE IF NOT EXISTS "sub" (
	"id" INTEGER NOT NULL,
	"enable" BOOLEAN NOT NULL DEFAULT true,
	"name" TEXT,
	"cron_expr" TEXT NOT NULL,
	"config" TEXT NOT NULL,
	"result" TEXT DEFAULT '{}',
	"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	"updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY("id")
);

CREATE TABLE IF NOT EXISTS "share" (
	"id" INTEGER NOT NULL,
	"enable" BOOLEAN NOT NULL DEFAULT false,
	"name" TEXT NOT NULL,
	"gen" TEXT NOT NULL,
	"token" TEXT NOT NULL UNIQUE,
	"access_count" INTEGER DEFAULT 0,
	"max_access_count" INTEGER DEFAULT 0,
	"expires" INTEGER DEFAULT 0,
	PRIMARY KEY("id")
);

CREATE TABLE IF NOT EXISTS "migration" (
	"date" INTEGER NOT NULL UNIQUE,
	"version" TEXT NOT NULL,
	"description" TEXT,
	"applied_at" DATETIME NOT NULL,
	PRIMARY KEY("date")
);
`
}

// init 自动注册迁移
func init() {
	migration.Register(ClientName, 202507171100, "dev", "Tables", Migration001Table)
}
