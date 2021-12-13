package main

import (
	"github.com/fzls/logger"
)

func initLogger() {
	logger.InitLogger("logs", "qinglong", "info")
}
