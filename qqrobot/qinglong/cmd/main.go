package main

import (
	logger "github.com/sirupsen/logrus"

	"github.com/Mrs4s/go-cqhttp/qqrobot/qinglong"
)

func main() {
	initLogger()

	envInfos, err := qinglong.ParseJdCookie()
	logger.Infof("my: %v %v", envInfos["jd_70a2bcede031c"], err)

	var info *qinglong.JdCookieInfo

	for _, param := range []string{"", "jd_70a2bcede031c", "1", "风之凌殇"} {
		info = qinglong.QueryCookieInfo(param)
		logger.Infof("%v: %v", param, info)
	}

	chart := qinglong.QueryChartPath(info)
	summary := qinglong.QuerySummary(info)
	logger.Info(chart)
	logger.Info(summary)
}
