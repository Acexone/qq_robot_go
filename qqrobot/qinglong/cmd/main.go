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
	for _, param := range []string{"", "jd_70a2bcede031c", "1", "风之凌殇"} {
		info = qinglong.QueryCookieInfo(param)
		logger.Infof("%v: %v", param, info)
	}

	chart := qinglong.QueryChartPath(info)
	summary := qinglong.QuerySummary(info)
	logger.Infof("")
	logger.Infof("解析的统计信息如下")
	logger.Infof("chart: %v", chart)
	logger.Infof("summary\n%v", summary)
}
