package storage

import (
	storageModel "github.com/evilCYH/NodeHub/internal/models/storage"
	"github.com/evilCYH/NodeHub/internal/modules/register"
	_ "github.com/evilCYH/NodeHub/internal/modules/storage/channel"
	"github.com/evilCYH/NodeHub/internal/utils/desc"
)

type Desc = desc.Data

func Get(m string, c string) (storageModel.Instance, error) {
	return register.Get[storageModel.Instance]("storage", m, c)
}

func GetChannels() []string {
	return register.GetList("storage")
}

func GetInfoMap() map[string][]Desc {
	return register.GetInfoMap("storage")
}
