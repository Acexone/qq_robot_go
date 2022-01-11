package qinglong

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
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

// JdCookieInfo 所需的京东cookie信息
type JdCookieInfo struct {
	PtPin  string // pt_pin, note: 定位只能使用这个字段，而不能使用index，因青龙不是依据env.sh来生成index的
	Remark string // remark
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
			PtPin:  ptPin,
			Remark: remark,
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
