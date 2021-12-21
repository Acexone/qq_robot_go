package qinglong

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	logger "github.com/sirupsen/logrus"

	"github.com/Mrs4s/go-cqhttp/global"
)

// maxCheckCount 通过搜索历史日志搜寻信息时，最多尝试的日志文件数目
const maxCheckCount = 6

// QueryCookieInfo 尝试通过pt_pin/序号/昵称等来查询cookie信息
func QueryCookieInfo(param string) *JdCookieInfo {
	if param == "" {
		return nil
	}

	ptPinToCookieInfo, err := ParseJdCookie()
	if err != nil {
		return nil
	}
	// 尝试pt_pin
	if info, ok := ptPinToCookieInfo[param]; ok {
		return info
	}

	// 尝试昵称
	for _, info := range ptPinToCookieInfo {
		remark := getRemark(info.Remark)
		if strings.Contains(remark, param) {
			return info
		}
	}

	return nil
}

// QueryChartPath 查询账号对应的统计图的路径
func QueryChartPath(info *JdCookieInfo) string {
	if info == nil {
		return ""
	}

	imageDir := getPath("log/.bean_chart")
	path, err := filepath.Abs(fmt.Sprintf("%s/chart_%v.jpeg", imageDir, info.PtPin))
	if err != nil {
		return ""
	}

	if !global.PathExists(path) {
		return ""
	}

	return path
}

// QuerySummary 查询账号对应的最新统计信息
func QuerySummary(info *JdCookieInfo) string {
	if info == nil {
		return ""
	}

	summaryDir := getPath("log/shufflewzc_faker2_jd_bean_change")
	logFiles, err := ioutil.ReadDir(summaryDir)
	if err != nil {
		logger.Errorf("read log dir failed, err=%v", err)
		return ""
	}

	// 按时间逆序排列
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].Name() > logFiles[j].Name()
	})

	// 因为有可能最新的日志还在处理中，因此逆序搜索一定数目的日志，直到搜索到为止
	for idx, logFile := range logFiles {
		if idx >= maxCheckCount {
			break
		}

		summary := parseSummary(info, filepath.Join(summaryDir, logFile.Name()))
		if summary != "" {
			return appendLogFileInfo(summary, logFile.Name())
		}
	}

	return ""
}

// QueryCookieExpired 查询账号是否已过期
func QueryCookieExpired(info *JdCookieInfo) string {
	if info == nil {
		return ""
	}

	checkCookieDir := getPath("log/shufflewzc_faker2_jd_CheckCK")
	logFiles, err := ioutil.ReadDir(checkCookieDir)
	if err != nil {
		logger.Errorf("read log dir failed, err=%v", err)
		return ""
	}

	if len(logFiles) == 0 {
		return ""
	}

	// 按时间逆序排列
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].Name() > logFiles[j].Name()
	})

	// 因为有可能最新的日志还在处理中，因此逆序搜索一定数目的日志，直到搜索到为止
	for idx, logFile := range logFiles {
		if idx >= maxCheckCount {
			break
		}

		result := parseCookieExpired(info, filepath.Join(checkCookieDir, logFile.Name()))
		if result != "" {
			return appendLogFileInfo(result, logFile.Name())
		}
	}

	return ""
}

func isCookieExpired(info *JdCookieInfo) bool {
	result := QueryCookieExpired(info)
	return strings.Contains(result, "已失效")
}

func appendLogFileInfo(parsedContents string, logFileName string) string {
	logTime := strings.TrimSuffix(logFileName, ".log")
	parsedTime, _ := time.Parse("2006-01-02-15-04-05", logTime)

	return fmt.Sprintf("%s\n\n-- 从 %s 更新的日志中解析得到", parsedContents, parsedTime.Format("2006-01-02 15:04:05"))
}
