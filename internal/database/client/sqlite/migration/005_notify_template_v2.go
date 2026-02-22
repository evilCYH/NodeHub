package migration

import "github.com/bestruirui/bestsub/internal/database/migration"

// Migration005NotifyTemplateV2 扩展通知模板字段（幂等）
func Migration005NotifyTemplateV2() string {
	return `
CREATE TABLE IF NOT EXISTS notify_template_v2 (
	"type" TEXT NOT NULL,
	"title" TEXT NOT NULL DEFAULT '',
	"content" TEXT NOT NULL DEFAULT '',
	PRIMARY KEY("type")
);

INSERT OR IGNORE INTO notify_template_v2 (type, title, content)
SELECT type, '', template FROM notify_template;

DROP TABLE IF EXISTS notify_template;
ALTER TABLE notify_template_v2 RENAME TO notify_template;
`
}

func init() {
	migration.Register(ClientName, 202512181000, "dev", "Notify Template V2", Migration005NotifyTemplateV2)
}
