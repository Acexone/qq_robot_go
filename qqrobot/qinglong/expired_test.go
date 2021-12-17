package qinglong

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseCookieExpired(t *testing.T) {
	logPath := getPath("log/shufflewzc_faker2_jd_CheckCK/2021-12-17-12-00-01.log")

	info := QueryCookieInfo("pin_1")
	assert.Equal(t, "pin_1 状态正常!", parseCookieExpired(info, logPath))

	info = QueryCookieInfo("pin_2")
	assert.Equal(t, "pin_2 已失效,自动禁用成功!", parseCookieExpired(info, logPath))

	info = QueryCookieInfo("pin_3")
	assert.Equal(t, "pin_3 已失效,已禁用!", parseCookieExpired(info, logPath))

	info = QueryCookieInfo(url.QueryEscape("中文pin"))
	assert.Equal(t, "中文pin 状态正常!", parseCookieExpired(info, logPath))
}
