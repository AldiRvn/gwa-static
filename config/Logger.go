package config

import (
	"log"
	"log/slog"
	"time"

	"github.com/logdyhq/logdy-core/logdy"
	"github.com/logdyhq/logdy-core/utils"
)

var Logger logdy.Logdy

func init() {
	appPort := "8081"

	Logger = logdy.InitializeLogdy(logdy.Config{
		ServerIp:   "0.0.0.0",
		ServerPort: appPort,
		LogInterceptor: func(entry *utils.LogEntry) {
			slog.Info("logdy access notif:", "time", entry.Time.Format(time.DateTime), "message", entry.Message)
		},
	}, nil)
	log.Printf("log web: http://0.0.0.0:%s\n", appPort)
}
