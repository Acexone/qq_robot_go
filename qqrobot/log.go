package qqrobot

import (
	"github.com/fzls/logger"
)

func initLogger() {
	logger.InitLogger("logs", "qq_robot", "debug")
}
