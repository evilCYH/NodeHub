package register

import (
	"github.com/evilCYH/NodeHub/internal/models/check"
	"github.com/evilCYH/NodeHub/internal/models/notify"
	"github.com/evilCYH/NodeHub/internal/models/storage"
)

func Notify(i notify.Instance) {
	register("notify", i)
}
func Check(i check.Instance) {
	register("check", i)
}
func Storage(i storage.Instance) {
	register("storage", i)
}
