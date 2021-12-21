package main

import (
	logger "github.com/sirupsen/logrus"

	"github.com/Mrs4s/go-cqhttp/qqrobot/qinglong"
)

func main() {
	initLogger()

	logger.Infof("解析cookie")
	envInfos, err := qinglong.ParseJdCookie()
	logger.Infof("my: %v %v", envInfos["jd_70a2bcede031c"], err)

	var info *qinglong.JdCookieInfo

	logger.Infof("")
	logger.Infof("使用不同方式查询cookie信息")
	for _, param := range []string{"", "pin_1", "测试账号-1"} {
		info = qinglong.QueryCookieInfo(param)
		logger.Infof("%v: %v", param, info)
	}

	chart := qinglong.QueryChartPath(info)
	summary := qinglong.QuerySummary(info)
	expired := qinglong.QueryCookieExpired(info)
	logger.Infof("")
	logger.Infof("解析的统计信息如下")
	logger.Infof("chart: %v", chart)
	logger.Infof("summary: \n%v\n", summary)
	logger.Infof("expired: \n%v\n", expired)
}
