package migration

import (
	"sort"
)

type Info struct {
	Date        uint64
	Version     string
	Description string
	Content     func() string
}

var clientMigrations = make(map[string][]*Info)

func Register(client string, date uint64, version, description string, contentFunc func() string) {
	info := &Info{
		Date:        date,
		Version:     version,
		Description: description,
		Content:     contentFunc,
	}

	migrations := clientMigrations[client]

	index := sort.Search(len(migrations), func(i int) bool {
		return migrations[i].Date > date
	})

	migrations = append(migrations, nil)
	copy(migrations[index+1:], migrations[index:])
	migrations[index] = info

	clientMigrations[client] = migrations
}

func Get(client string) []*Info {
	if migrations := clientMigrations[client]; migrations != nil {
		result := make([]*Info, len(migrations))
		copy(result, migrations)
		return result
	}
	return make([]*Info, 0)
}
