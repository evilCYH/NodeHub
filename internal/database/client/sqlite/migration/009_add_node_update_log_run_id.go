package migration

import "github.com/evilCYH/NodeHub/internal/database/migration"

// Migration009AddNodeUpdateLogRunID 为 node_update_log 增加 run_id 关联字段
func Migration009AddNodeUpdateLogRunID() string {
	return `
ALTER TABLE node_update_log ADD COLUMN run_id INTEGER DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_node_update_log_sub_run ON node_update_log(sub_id, run_id);
`
}

func init() {
	migration.Register(ClientName, 202602271000, "dev", "Add Node Update Log Run ID", Migration009AddNodeUpdateLogRunID)
}
