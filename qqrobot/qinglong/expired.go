package qinglong

import (
	"io/ioutil"
	"strings"

	"github.com/Mrs4s/go-cqhttp/global"
)

func parseCookieExpired(info *JdCookieInfo, logFilePath string) string {
	if !global.PathExists(logFilePath) {
		return ""
	}

	contentBytes, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		return ""
	}
	content := string(contentBytes)

	prefixToRemove := " : "
	prefix := prefixToRemove + info.QueryUnescapedPtPin()
	suffix := "\n\n"

	// 定位前缀
	prefixIndex := strings.Index(content, prefix)
	if prefixIndex == -1 {
		return ""
	}
	prefixIndex += len(prefixToRemove)

	// 定位后缀
	suffixIndex := strings.Index(content[prefixIndex:], suffix)
	if suffixIndex == -1 {
		return ""
	}
	suffixIndex += prefixIndex

	result := content[prefixIndex:suffixIndex]

	return result
}
