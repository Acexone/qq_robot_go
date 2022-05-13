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

// maxExpireCheckCount 由于可能最新的还在处理中，因此最多尝试两个，避免过多查询导致之前的过期记录也影响这个
const maxExpireCheckCount = 2

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

// QuerySummary 查询账号对应的最新统计信息，返回第一个包含农场信息的结果，如果都不包含，则返回最后一个不为空的结果
func QuerySummary(info *JdCookieInfo) string {
	if info == nil {
		return ""
	}

	targetSummary := ""

	logFileList := []string{
		"log/ccwav_QLScript2_jd_bean_change",
		"log/KingRan_KR_jd_bean_change_pro",
	}
	for _, logFile := range logFileList {
		summary := querySummary(info, logFile)
		if summary != "" {
			targetSummary = summary
		}
		if strings.Contains(summary, "东东农场") {
			// 从多个日志来源中查询概览，优先包含 东东农场 信息的那个，因为有时候可能其中一个接口会无法查询农场的信息
			break
		}
	}

	return targetSummary
}

func querySummary(info *JdCookieInfo, logFile string) string {
	summaryDir := getPath(logFile)
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

	checkCookieDir := getPath("log/ccwav_QLScript2_jd_CheckCK")
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
	processedValidLogCount := 0
	for _, logFile := range logFiles {
		result, isLogComplete, skipThis := parseCookieExpired(info, filepath.Join(checkCookieDir, logFile.Name()))
		if result != "" {
			return appendLogFileInfo(result, logFile.Name())
		} else if isLogComplete {
			// 没有解析到该账号，但该日志是完整日志，说明这个账号未过期
			return ""
		} else if skipThis {
			// 如果内部判定需要跳过该文件，则不计数，继续尝试下一个
		} else {
			processedValidLogCount++
		}

		// 仅尝试该数目个正常的日志
		if processedValidLogCount >= maxExpireCheckCount {
			break
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

	return fmt.Sprintf("%s\n【解析于 %s 的日志】", parsedContents, parsedTime.Format("2006-01-02 15:04:05"))
}
