package migration

import "github.com/evilCYH/NodeHub/internal/database/migration"

const ClientName = "sqlite"

func Get() []*migration.Info {
	return migration.Get(ClientName)
}
