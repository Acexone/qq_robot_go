package qinglong

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Mrs4s/go-cqhttp/global"
)

func parseSummary(info *JdCookieInfo, logFilePath string) string {
	if !global.PathExists(logFilePath) {
		return ""
	}

	contentBytes, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		return ""
	}
	content := string(contentBytes)

	blockPrefix := fmt.Sprintf("】%v*********", info.QueryUnescapedPtPin())
	prefix := "【账号"
	suffix := "\n\n"

	// 定位唯一区分账号的日志前缀
	blockPrefixIndex := strings.Index(content, blockPrefix)
	if blockPrefixIndex == -1 {
		return ""
	}

	// 定位实际前缀
	prefixIndex := strings.Index(content[blockPrefixIndex:], prefix)
	if prefixIndex == -1 {
		return ""
	}
	prefixIndex += blockPrefixIndex

	// 定位后缀
	suffixIndex := strings.Index(content[prefixIndex:], suffix)
	if suffixIndex == -1 {
		return ""
	}
	suffixIndex += prefixIndex

	summary := content[prefixIndex:suffixIndex]

	return summary
}
