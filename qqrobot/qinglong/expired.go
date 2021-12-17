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
	prefix := prefixToRemove + info.PtPin
	suffix := "\n\n"

	prefixIndex := strings.Index(content, prefix)
	if prefixIndex == -1 {
		return ""
	}
	relativeSuffixIndex := strings.Index(content[prefixIndex:], suffix)
	if relativeSuffixIndex == -1 {
		return ""
	}
	suffixIndex := prefixIndex + relativeSuffixIndex

	result := content[prefixIndex+len(prefixToRemove) : suffixIndex]

	return result
}
