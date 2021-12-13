package qinglong

import (
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
