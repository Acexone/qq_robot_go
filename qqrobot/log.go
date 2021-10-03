package qqrobot

import (
	"fmt"

	mylogger "github.com/fzls/logger"
)

var logger *mylogger.SugaredLogger

func init() {
	// 初始化日志
	var err error
	logger, err = mylogger.NewLogger("logs", "qq_robot", mylogger.InfoLevel)
	if err != nil {
		fmt.Printf("new logger err=%v\n", err)
		return
	}
}
