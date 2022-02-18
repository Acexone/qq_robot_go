package qinglong

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

// EnvDBEntry env.db中记录格式
type EnvDBEntry struct {
	Value     string  `json:"value"`
	ID        string  `json:"_id"`
	Created   int64   `json:"created"`
	Status    int     `json:"status"`
	Timestamp string  `json:"timestamp"`
	Position  float64 `json:"position"`
	Name      string  `json:"name"`
	Remarks   string  `json:"remarks"`
}

// CreateTime 创建unix时间戳
func (e *EnvDBEntry) CreateTime() int64 {
	return time.UnixMilli(e.Created).Unix()
}

// UpdateTime 更新unix时间戳
func (e *EnvDBEntry) UpdateTime() int64 {
	// re: 鉴于nvjdc经常不可用，避免手动更新后不会更新该时间戳，暂时屏蔽
	// // 先尝试解析nvjdc设置的更新时间：test@@1640780099690@@UID_xxxxxx
	// if strings.Contains(e.Remarks, "@@") {
	// 	for _, remarkPart := range strings.Split(e.Remarks, "@@") {
	// 		if len(remarkPart) != 13 {
	// 			continue
	// 		}
	//
	// 		millTimeStamp, err := strconv.ParseInt(remarkPart, 10, 64)
	// 		if err != nil {
	// 			continue
	// 		}
	//
	// 		return time.UnixMilli(millTimeStamp).Unix()
	// 	}
	// }

	// 否则，解析青龙设置的更新时间
	// Wed Nov 10 2021 17:28:28 GMT+0800 (中国标准时间)
	t, _ := time.Parse("Mon Jan 02 2006 15:04:05 MST-0700 (中国标准时间)", e.Timestamp)
	return t.Unix()
}

// UsedDays 已使用天数
func (e *EnvDBEntry) UsedDays() int {
	return int(time.Now().Unix()-e.CreateTime()) / 86400
}

const maxValidDays = 30

// EstimateRemainingDays 预计剩余天数
func (e *EnvDBEntry) EstimateRemainingDays() int {
	return maxValidDays - int(time.Now().Unix()-e.UpdateTime())/86400
}

// JdCookieInfo 所需的京东cookie信息
type JdCookieInfo struct {
	PtPin                 string // pt_pin, note: 定位只能使用这个字段，而不能使用index，因青龙不是依据env.sh来生成index的
	Remark                string // remark
	UsedDays              int    // 已使用天数
	EstimateRemainingDays int    // 预计剩余天数
}

// QueryUnescapedPtPin 返回url解码后的pt_pin
func (info *JdCookieInfo) QueryUnescapedPtPin() string {
	// 由于可能部分pt_pin包含中文的url编码，而日志中均会打印为解码后的，需要解码一次
	pin, err := url.QueryUnescape(info.PtPin)
	if err != nil {
		return info.PtPin
	}

	return pin
}

// ToChatMessage 转换为聊天消息
func (info *JdCookieInfo) ToChatMessage() string {
	expiredInfo := "未过期"
	// 检查是否过期
	if isCookieExpired(info) {
		expiredInfo = "已过期，请更新cookie（每六个小时重新检测并自动启用）"
	} else if info.UsedDays != 0 {
		expiredInfo += fmt.Sprintf("，已使用 %d 天, 预计 %d 天后过期", info.UsedDays, info.EstimateRemainingDays)
	}
	return fmt.Sprintf("\n"+
		"\npt_pin: %v"+
		"\n备注: %v"+
		"\n状态: %v"+
		"",
		info.PtPin,
		info.Remark,
		expiredInfo,
	)
}

// ParseJdCookie 解析 db/env.db 和 config/env.sh，获取各个京东cookie的序号和备注信息
func ParseJdCookie() (map[string]*JdCookieInfo, error) {
	ptPinToCookieInfo, err := parseEnvDB()
	if err != nil {
		return nil, err
	}

	ptPinToIndex, err := parseEnvSh()
	if err != nil {
		return nil, err
	}
	for ptPin := range ptPinToIndex {
		cookieInfo := ptPinToCookieInfo[ptPin]
		if cookieInfo == nil {
			// 部分账号可能在env.db中不存在，但是env.sh中有
			ptPinToCookieInfo[ptPin] = &JdCookieInfo{
				PtPin:  ptPin,
				Remark: ptPin,
			}
		}
	}

	return ptPinToCookieInfo, nil
}

func parseEnvDB() (map[string]*JdCookieInfo, error) {
	envDBPath := getPath("db/env.db")

	envDBFile, err := os.Open(envDBPath)
	if err != nil {
		return nil, err
	}
	defer envDBFile.Close()

	ptPinToCookieInfo := make(map[string]*JdCookieInfo)

	scanner := bufio.NewScanner(envDBFile)
	for scanner.Scan() {
		var envEntry EnvDBEntry
		err := json.Unmarshal(scanner.Bytes(), &envEntry)
		if err != nil {
			return nil, err
		}

		if envEntry.Name != "JD_COOKIE" {
			continue
		}

		ptPin := getPtPin(envEntry.Value)
		remark := getRemark(envEntry.Remarks)

		ptPinToCookieInfo[ptPin] = &JdCookieInfo{
			PtPin:                 ptPin,
			Remark:                remark,
			UsedDays:              envEntry.UsedDays(),
			EstimateRemainingDays: envEntry.EstimateRemainingDays(),
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ptPinToCookieInfo, nil
}

func parseEnvSh() (map[string]int, error) {
	envShPath := getPath("config/env.sh")

	envShFile, err := os.Open(envShPath)
	if err != nil {
		return nil, err
	}
	defer envShFile.Close()

	ptPinToIndex := make(map[string]int)

	scanner := bufio.NewScanner(envShFile)
	for scanner.Scan() {
		// export ENV_VAR="XXXX"
		envDefineStatement := scanner.Text()
		if !strings.HasPrefix(envDefineStatement, `export JD_COOKIE="`) {
			continue
		}

		cookieStrings := envDefineStatement[len(`export JD_COOKIE="`) : len(envDefineStatement)-1]
		for index, cookie := range strings.Split(cookieStrings, "&") {
			ptPin := getPtPin(cookie)
			ptPinToIndex[ptPin] = index + 1
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ptPinToIndex, nil
}

// 以下函数针对的cookie形如 pt_key=XXXX;pt_pin=YYYY;

func getPtPin(cookie string) string {
	return getCookie(cookie, "pt_pin")
}

func getCookie(cookie string, cookieKey string) string {
	for _, kv := range strings.Split(cookie, ";") {
		kvArr := strings.SplitN(kv, "=", 2)
		if len(kvArr) != 2 {
			continue
		}
		if kvArr[0] == cookieKey {
			return kvArr[1]
		}
	}

	return ""
}

func getRemark(remarks string) string {
	// 如果有备注，则使用备注
	remark := remarks

	// 如果备注中有remark=字样
	for _, remarkPart := range strings.Split(remarks, ";") {
		if strings.HasPrefix(remarkPart, "remark=") {
			// remark=备注;
			remark = strings.TrimPrefix(remarkPart, "remark=")
			break
		}
	}

	// 特殊处理nvdjc附加的备注
	//   remark=test;@@UID_xxxxxx
	//   test@@UID_xxxxxx
	//   test@@1640780099690@@UID_xxxxxx
	if strings.Contains(remark, "@@") {
		remark = strings.Split(remark, "@@")[0]
	}

	return remark
}
