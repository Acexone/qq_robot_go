package qinglong

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 2021/12/13 20:48 by fzls

func Test_parseSummary(t *testing.T) {
	logPath := getPath("log/KingRan_KR_jd_bean_change_pro/2022-04-21-21-30-00.log")

	info := QueryCookieInfo("pin_1")
	assert.Contains(t, parseSummary(info, logPath), "测试账号-1")

	info = QueryCookieInfo(url.QueryEscape("中文pin"))
	assert.Contains(t, parseSummary(info, logPath), "中文名字")

	info = QueryCookieInfo("pin_3")
	assert.Empty(t, parseSummary(info, logPath))
}
