package qinglong

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// QueryCookieInfo 尝试通过pt_pin/序号/昵称等来查询cookie信息
func QueryCookieInfo(param string) *JdCookieInfo {
	if param == "" {
		return nil
	}

	ptPinToCookieInfo, err := ParseJdCookie()
	if err != nil {
		return nil
	}
	// 尝试pt_pin
	if info, ok := ptPinToCookieInfo[param]; ok {
		return info
	}

	// 尝试序号
	index, _ := strconv.ParseInt(param, 10, 64)
	for _, info := range ptPinToCookieInfo {
		if info.Index == int(index) {
			return info
		}
	}

	// 尝试昵称
	for _, info := range ptPinToCookieInfo {
		remark := getRemark(info.Remark)
		if strings.Contains(remark, param) {
			return info
		}
	}

	return nil
}

// QueryChartPath 查询账号对应的统计图的路径
func QueryChartPath(info *JdCookieInfo) string {
	if info == nil {
		return ""
	}

	imageDir := getPath("log/.bean_chart")
	path, err := filepath.Abs(fmt.Sprintf("%s/chart_%v.jpeg", imageDir, info.Index))
	if err != nil {
		return ""
	}

	return path
}
