package qinglong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 2021/12/13 20:48 by fzls

func Test_parseSummary(t *testing.T) {
	logPath := getPath("log/shufflewzc_faker2_jd_bean_change/2021-12-13-09-30-00.log")

	info := QueryCookieInfo("1")
	assert.NotEmpty(t, parseSummary(info, logPath))

	info = QueryCookieInfo("3")
	assert.Empty(t, parseSummary(info, logPath))
}
