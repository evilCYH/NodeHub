package shutdown

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/evilCYH/NodeHub/internal/utils/log"
)

var funcs []func() error

func Register(fn func() error) {
	funcs = append(funcs, fn)
}

func Listen() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	log.Info("Program started, press Ctrl+C to exit")
	sig := <-quit
	log.Warnf("Received exit signal: %v, starting to close program", sig)
	if len(funcs) == 0 {
		return
	}
	for _, fn := range funcs {
		if err := fn(); err != nil {
			log.Errorf("Closing functions execution failed: %v", err)
		}
	}
	log.Info("=== Shutdown completed successfully ===")
	os.Exit(0)
}
func All() {
	if len(funcs) == 0 {
		return
	}
	for _, fn := range funcs {
		if err := fn(); err != nil {
			log.Errorf("Closing functions execution failed: %v", err)
		}
	}
	log.Info("Shutdown completed successfully")
}
