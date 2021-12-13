package main

import (
	logger "github.com/sirupsen/logrus"

	"github.com/Mrs4s/go-cqhttp/qqrobot/qinglong"
)

func main() {
	initLogger()

	envInfos, _ := qinglong.ParseJdCookie()
	logger.Infof("my: %v", envInfos["jd_70a2bcede031c"])

	var info *qinglong.JdCookieInfo

	for _, param := range []string{"", "jd_70a2bcede031c", "1", "风之凌殇"} {
		info = qinglong.QueryCookieInfo(param)
		chart := qinglong.QueryChartPath(info)

		// logger.Infof("%v: %v", param, info)
		logger.Infof("%v: %v", param, chart)
	}
}
