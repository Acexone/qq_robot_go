package qinglong

import (
	"bufio"
	"encoding/json"
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
	Index  int    // 在env.sh中的JD_COOKIE中的顺序，从1开始计数
	PtPin  string // pt_pin
	Remark string // remark
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
	for ptPin, index := range ptPinToIndex {
		cookieInfo := ptPinToCookieInfo[ptPin]
		if cookieInfo != nil {
			cookieInfo.Index = index
		} else {
			// 部分账号可能在env.db中不存在，但是env.sh中有
			ptPinToCookieInfo[ptPin] = &JdCookieInfo{
				Index:  index,
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
			Index:  -1,
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
	if strings.HasPrefix(remarks, "remark=") {
		// remark=备注;
		remarks = strings.TrimPrefix(remarks, "remark=")
		remarks = strings.TrimSuffix(remarks, ";")
	}

	return remarks
}
