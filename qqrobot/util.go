package qqrobot

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/internal/download"
	"github.com/gookit/color"
	logger "github.com/sirupsen/logrus"

	"github.com/Mrs4s/go-cqhttp/coolq"
	"github.com/Mrs4s/go-cqhttp/global"
)

const (
	maxImageSize = 1024 * 1024 * 30 // 30MB
)

func (r *QQRobot) tryAppendImageByURL(m *message.SendingMessage, imageURL string) {
	image, err := r._makeLocalImage(imageURL)
	if err != nil {
		logger.Errorf("_makeLocalImage err=%v", err)
		return
	}

	m.Append(image)
}

// modified based on makeImageOrVideoElem
func (r *QQRobot) _makeLocalImage(imageURL string) (message.IMessageElement, error) {
	hash := md5.Sum([]byte(imageURL))
	cacheFile := path.Join(global.CachePath, hex.EncodeToString(hash[:])+".cache")
	maxSize := func() int64 {
		return maxImageSize
	}()
	exist := global.PathExists(cacheFile)
	if exist {
		goto hasCacheFile
	}
	{
		req := download.Request{URL: imageURL, Limit: maxSize}
		if err := req.WriteToFileMultiThreading(cacheFile, runtime.NumCPU()); err != nil {
			return nil, err
		}
	}

hasCacheFile:
	return &coolq.LocalImageElement{File: cacheFile}, nil
}

func (r *QQRobot) ocr(groupImageElement *message.GroupImageElement) (ocrResultString string) {
	imageMd5 := fmt.Sprintf("%x", groupImageElement.Md5)

	cached, ok := r.ocrCache.Get(imageMd5)
	if ok {
		return cached.(string)
	}

	defer func() {
		r.ocrCache.Add(imageMd5, ocrResultString)
	}()

	ocrResult, err := r.cqBot.Client.ImageOcr(groupImageElement)
	if err != nil {
		logger.Errorf("ocr出错了，image=%+v，err=%v", groupImageElement.Url, err)
		return ""
	}

	resultBuffer := strings.Builder{}
	for _, textDetection := range ocrResult.Texts {
		resultBuffer.WriteString(textDetection.Text)
	}
	ocrResultString = resultBuffer.String()

	logger.Infof(bold(color.Yellow).Render(fmt.Sprintf("ocr ok image=%v  result is:\n%v", groupImageElement.Url, ocrResultString)))
	return ocrResultString
}

func (r *QQRobot) sendTextMessageToGroup(groupID int64, msg string) {
	r.cqBot.SendGroupMessage(groupID, message.NewSendingMessage().Append(message.NewText(msg)))
}

func (r *QQRobot) currentTime() string {
	return r.formatTime(time.Now())
}

func (r *QQRobot) formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.999")
}

func getCurrentPeriodName() string {
	hour := time.Now().Hour()
	switch {
	case hour == 23 || 0 <= hour && hour < 5:
		return "深夜"
	case 5 <= hour && hour < 11:
		return "早上"
	case 11 <= hour && hour < 13:
		return "中午"
	case 13 <= hour && hour < 17:
		return "下午"
	default:
		return "晚上"
	}
}

// 是否是管理员
func isMemberAdmin(permission client.MemberPermission) bool {
	return permission == client.Owner || permission == client.Administrator
}

// 单条消息发送的大小有限制，所以需要分成多段来发
const maxMessageSize = 5000

func splitPlainMessage(content string) []message.IMessageElement {
	if len(content) <= maxMessageSize {
		return []message.IMessageElement{message.NewText(content)}
	}

	var splittedMessage []message.IMessageElement

	var part string
	remainingText := content
	for len(remainingText) != 0 {
		partSize := 0
		for _, runeValue := range remainingText {
			runeSize := len(string(runeValue))
			if partSize+runeSize > maxMessageSize {
				break
			}
			partSize += runeSize
		}

		part, remainingText = remainingText[:partSize], remainingText[partSize:]
		splittedMessage = append(splittedMessage, message.NewText(part))
	}

	return splittedMessage
}

func p(v interface{}) string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
}
