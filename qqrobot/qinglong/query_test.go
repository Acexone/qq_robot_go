package qinglong

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 2021/12/13 20:33 by fzls

func TestQueryCookieInfo(t *testing.T) {
	var info *JdCookieInfo

	// 无参数
	info = QueryCookieInfo("")
	assert.Nil(t, info)

	// pin
	info = QueryCookieInfo("pin_1")
	assert.NotNil(t, info)

	// 序号
	info = QueryCookieInfo("1")
	assert.NotNil(t, info)

	// 备注
	info = QueryCookieInfo("测试账号-1")
	assert.NotNil(t, info)

	// 仅env中存在的账号，使用pin
	info = QueryCookieInfo("pin_3")
	assert.NotNil(t, info)

	// 不存在的账号
	info = QueryCookieInfo("not exists")
	assert.Nil(t, info)
}

func TestQueryChartPath(t *testing.T) {
	info := QueryCookieInfo("1")
	chartPath := QueryChartPath(info)
	expected, _ := filepath.Abs(getPath("log/.bean_chart/chart_pin_1.jpeg"))
	assert.Equal(t, expected, chartPath)
}

func TestQuerySummary(t *testing.T) {
	info := QueryCookieInfo("1")
	assert.NotEmpty(t, QuerySummary(info))

	info = QueryCookieInfo("3")
	assert.Empty(t, QuerySummary(info))
}

func TestQueryCookieExpired(t *testing.T) {
	info := QueryCookieInfo("1")
	assert.NotEmpty(t, QueryCookieExpired(info))

	info = QueryCookieInfo("2")
	assert.NotEmpty(t, QueryCookieExpired(info))

	info = QueryCookieInfo("3")
	assert.NotEmpty(t, QueryCookieExpired(info))

	info = QueryCookieInfo("4")
	assert.NotEmpty(t, QueryCookieExpired(info))

	info = QueryCookieInfo("99999")
	assert.Empty(t, QueryCookieExpired(info))
}
