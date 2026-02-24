package check

import (
	_ "github.com/evilCYH/NodeHub/internal/core/check/checker"
	"github.com/evilCYH/NodeHub/internal/models/check"
	"github.com/evilCYH/NodeHub/internal/modules/register"
	"github.com/evilCYH/NodeHub/internal/utils/desc"
)

type Desc = desc.Data

func Get(m string, c string) (check.Instance, error) {
	return register.Get[check.Instance]("check", m, c)
}

func GetTypes() []string {
	return register.GetList("check")
}

func GetInfoMap() map[string][]Desc {
	return register.GetInfoMap("check")
}
