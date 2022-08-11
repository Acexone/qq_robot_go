package qinglong

import (
	"os"
	"strings"

	"github.com/Mrs4s/go-cqhttp/global"
)

func parseCookieExpired(info *JdCookieInfo, logFilePath string) (result string, isLogComplete bool, skipThis bool) {
	if !global.PathExists(logFilePath) {
		return "", false, true
	}

	contentBytes, err := os.ReadFile(logFilePath)
	if err != nil {
		return "", false, true
	}
	content := string(contentBytes)

	// 跳过未实际执行的日志
	// Error: Cannot find module '/ql/scripts/ccwav_QLScript2_jd_CheckCK.js'
	if strings.Contains(content, "Error: Cannot find module") {
		return "", false, true
	}

	// 判断这个日志是否运行完整
	isLogComplete = strings.Contains(content, "开始发送通知...")

	prefixToRemove := " : "
	prefix := prefixToRemove + info.QueryUnescapedPtPin()
	suffix := "\n\n"

	// 定位前缀
	prefixIndex := strings.Index(content, prefix)
	if prefixIndex == -1 {
		return "", isLogComplete, false
	}
	prefixIndex += len(prefixToRemove)

	// 定位后缀
	suffixIndex := strings.Index(content[prefixIndex:], suffix)
	if suffixIndex == -1 {
		return "", isLogComplete, false
	}
	suffixIndex += prefixIndex

	result = content[prefixIndex:suffixIndex]

	return result, isLogComplete, false
}
