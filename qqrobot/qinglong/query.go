package qinglong

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	logger "github.com/sirupsen/logrus"

	"github.com/Mrs4s/go-cqhttp/global"
)

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

	// 尝试序号
	index, _ := strconv.ParseInt(param, 10, 64)
	for _, info := range ptPinToCookieInfo {
		if info.Index == int(index) {
			return info
		}
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

	// 因为有可能最新的日志还在处理中，因此逆序搜索每一个日志，直到搜索到为止
	for _, logFile := range logFiles {
		summary := parseSummary(info, filepath.Join(summaryDir, logFile.Name()))
		if summary != "" {
			return summary
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

	// 只取最新的一个日志
	latestLogFile := logFiles[0]
	result := parseCookieExpired(info, filepath.Join(checkCookieDir, latestLogFile.Name()))
	if result != "" {
		return result
	}

	return ""
}
