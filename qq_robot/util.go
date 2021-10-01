package qq_robot

import (
	"crypto/md5"
	"encoding/hex"
	"path"
	"runtime"

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
