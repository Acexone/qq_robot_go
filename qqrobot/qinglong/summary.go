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

	prefix := fmt.Sprintf("【账号%v🆔】", info.Index)
	suffix := "\n\n"

	prefixIndex := strings.Index(content, prefix)
	if prefixIndex == -1 {
		return ""
	}
	suffixIndex := prefixIndex + strings.Index(content[prefixIndex:], suffix)

	summary := content[prefixIndex:suffixIndex]

	return summary
}
