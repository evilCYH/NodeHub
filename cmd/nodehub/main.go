package main

import (
	"flag"

	"github.com/evilCYH/NodeHub/internal/config"
	"github.com/evilCYH/NodeHub/internal/core/cron"
	"github.com/evilCYH/NodeHub/internal/core/node"
	"github.com/evilCYH/NodeHub/internal/core/task"
	"github.com/evilCYH/NodeHub/internal/database"
	"github.com/evilCYH/NodeHub/internal/database/op"
	"github.com/evilCYH/NodeHub/internal/models/setting"
	"github.com/evilCYH/NodeHub/internal/server/server"
	"github.com/evilCYH/NodeHub/internal/utils/info"
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"github.com/evilCYH/NodeHub/internal/utils/shutdown"
)

func main() {
	configPath := flag.String("c", "", "config file path")
	flag.Parse()
	config.MustLoad(*configPath)

	info.Banner()

	cfg := config.Base()

	if err := log.Initialize(cfg.Log.Level, cfg.Log.Path, cfg.Log.Output); err != nil {
		panic(err)
	}
	if err := database.Initialize(cfg.Database.Type, cfg.Database.Path); err != nil {
		panic(err)
	}

	if err := server.Initialize(); err != nil {
		panic(err)
	}

	task.Init(op.GetSettingInt(setting.TASK_MAX_THREAD))

	cron.Start()
	cron.FetchLoad()
	cron.CheckLoad()

	node.InitNodePool(op.GetSettingInt(setting.NODE_POOL_SIZE))

	server.Start()

	shutdown.Register(server.Close)       //   ↓↓
	shutdown.Register(database.Close)     //   ↓↓
	shutdown.Register(node.CloseNodePool) //   ↓↓
	shutdown.Register(log.Close)          //   ↓↓

	shutdown.Listen()
}
