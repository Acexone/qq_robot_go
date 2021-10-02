package qq_robot

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
	"github.com/Mrs4s/go-cqhttp/coolq"
	"github.com/Mrs4s/go-cqhttp/global"
	log "github.com/sirupsen/logrus"
)

const (
	maxImageSize = 1024 * 1024 * 30 // 30MB
)

func (r *QQRobot) tryAppendImageByUrl(m *message.SendingMessage, imageUrl string) {
	image, err := r._makeLocalImage(imageUrl)
	if err != nil {
		log.Errorf("_makeLocalImage err=%v", err)
		return
	}

	m.Append(image)
}

// modified based on makeImageOrVideoElem
func (r *QQRobot) _makeLocalImage(imageUrl string) (message.IMessageElement, error) {
	hash := md5.Sum([]byte(imageUrl))
	cacheFile := path.Join(global.CachePath, hex.EncodeToString(hash[:])+".cache")
	maxSize := func() int64 {
		return maxImageSize
	}()
	exist := global.PathExists(cacheFile)
	if exist {
		goto hasCacheFile
	}
	if err := global.DownloadFileMultiThreading(imageUrl, cacheFile, maxSize, runtime.NumCPU(), nil); err != nil {
		return nil, err
	}
hasCacheFile:
	return &coolq.LocalImageElement{File: cacheFile}, nil
}

func (r *QQRobot) ocr(groupImageElement *message.GroupImageElement) (ocrResultString string) {
	image_md5 := fmt.Sprintf("%x", groupImageElement.Md5)

	cached, ok := r.ocrCache.Get(image_md5)
	if ok {
		return cached.(string)
	}

	defer func() {
		r.ocrCache.Add(image_md5, ocrResultString)
	}()
	
	ocrResult, err := r.cqBot.Client.ImageOcr(groupImageElement)
	if err != nil {
		logger.Errorf("ocr出错了，image=%+v，err=%v", groupImageElement, err)
		return ""
	}

	resultBuffer := strings.Builder{}
	for _, textDetection := range ocrResult.Texts {
		resultBuffer.WriteString(textDetection.Text)
	}
	return resultBuffer.String()
}

func (r *QQRobot) currentTime() string {
	return r.formatTime(time.Now())
}

func (r *QQRobot) formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.999")
}

func getCurrentPeriodName() string {
	hour := time.Now().Hour()
	if hour == 23 || 0 <= hour && hour < 5 {
		return "深夜"
	} else if 5 <= hour && hour < 11 {
		return "早上"
	} else if 11 <= hour && hour < 13 {
		return "中午"
	} else if 13 <= hour && hour < 17 {
		return "下午"
	} else {
		return "晚上"
	}
}

// 是否是管理员
func isMemberAdmin(permission client.MemberPermission) bool {
	return permission == client.Owner || permission == client.Administrator
}

// 单条消息发送的大小有限制，所以需要分成多段来发
const maxMessageJsonSize = 400

func splitPlainMessage(content string) []message.IMessageElement {
	if len(content) <= maxMessageJsonSize {
		return []message.IMessageElement{message.NewText(content)}
	}

	var splittedMessage []message.IMessageElement

	var part string
	remainingText := content
	for len(remainingText) != 0 {
		partSize := 0
		for byteIdx, runeValue := range remainingText {
			if partSize+byteIdx > maxMessageJsonSize {
				break
			}
			partSize += len(string(runeValue))
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
