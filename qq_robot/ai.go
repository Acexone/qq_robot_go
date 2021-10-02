package qq_robot

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 2021/10/02 5:20 by fzls
type AiChatResponse struct {
	RetCode int                `json:"ret"`
	Message string             `json:"msg"`
	Data    AiChatResponseData `json:"data"`
}

type AiChatResponseData struct {
	Session string `json:"session"`
	Answer  string `json:"answer"`
}

// 使用腾讯ai开放平台的智能闲聊接口 https://ai.qq.com/doc/nlpchat.shtml
func (r *QQRobot) aiChat(targetQQ int64, chatText string) (responseText string) {
	cfg := r.Config.Robot

	params := url.Values{}
	params.Set("app_id", cfg.TencentAiAppId)
	params.Set("time_stamp", strconv.FormatInt(time.Now().Unix(), 10))
	params.Set("nonce_str", strconv.FormatInt(rand.Int63(), 10))
	params.Set("session", strconv.FormatInt(targetQQ, 10))
	params.Set("question", chatText)

	params.Set("sign", MakeSign(params, cfg.TencentAiAppKey))

	resp, err := r.HttpClient.PostForm(TencentAiApi, params)
	if err != nil {
		logger.Debugf("aiChat(qq=%v, text=%v) params=%v post err=%v", targetQQ, chatText, params.Encode(), err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger.Debugf("aiChat(qq=%v, text=%v) params=%v status code=%v", targetQQ, chatText, params.Encode(), resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Debugf("aiChat(qq=%v, text=%v) params=%v read body err=%v", targetQQ, chatText, params.Encode(), err)
		return
	}

	var result AiChatResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		logger.Debugf("aiChat(qq=%v, text=%v) params=%v unmarshal err=%v body=%v", targetQQ, chatText, params.Encode(), err, body)
		return
	}
	if result.RetCode != 0 {
		logger.Debugf("aiChat(qq=%v, text=%v) params=%v retcode!=0, result=%v", targetQQ, chatText, params.Encode(), result)
		return result.Message
	}

	return result.Data.Answer
}

func MakeSign(params url.Values, appKey string) string {
	// 将<key, value>请求参数对按key进行字典升序排序，得到有序的参数对列表N
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 将列表N中的参数对按URL键值对的格式拼接成字符串，得到字符串T（如：key1=value1&key2=value2），URL键值拼接过程value部分需要URL编码，URL编码算法用大写字母，例如%E8，而不是小写%e8
	var kvs []string
	for _, k := range keys {
		vs := params[k]
		for _, v := range vs {
			kvs = append(kvs, fmt.Sprintf("%v=%v", k, url.QueryEscape(v)))
		}
	}

	// 将应用密钥以app_key为键名，组成URL键值拼接到字符串T末尾，得到字符串S（如：key1=value1&key2=value2&app_key=密钥)
	kvs = append(kvs, fmt.Sprintf("%v=%v", "app_key", appKey))

	strToSign := strings.Join(kvs, "&")

	// 对拼接的数据库进行md5摘要，即可得sign签名
	w := md5.New()
	io.WriteString(w, strToSign)
	sign := fmt.Sprintf("%X", w.Sum(nil))

	return sign
}
