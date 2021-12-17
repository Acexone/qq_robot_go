package qinglong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseCookieExpired(t *testing.T) {
	logPath := getPath("log/shufflewzc_faker2_jd_CheckCK/2021-12-17-12-00-01.log")

	info := QueryCookieInfo("1")
	assert.Equal(t, "京东账号1 : pin_1 状态正常!", parseCookieExpired(info, logPath))

	info = QueryCookieInfo("2")
	assert.Equal(t, "京东账号2 : pin_2 已失效,自动禁用成功!", parseCookieExpired(info, logPath))

	info = QueryCookieInfo("3")
	assert.Equal(t, "京东账号3 : pin_3 已失效,已禁用!", parseCookieExpired(info, logPath))
}
