package qq_robot

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/global"
	"github.com/tidwall/gjson"
)

func (r *QQRobot) tryAppendMusicElement(m *message.SendingMessage, musicName string) {
	musicElem, err := r.makeMusicShareElement(musicName)
	if err != nil {
		return
	}

	m.Append(musicElem)
}

func (r *QQRobot) makeMusicShareElement(musicName string) (*message.MusicShareElement, error) {
	songId := strconv.FormatInt(r.queryNeteaseMusic(musicName), 10)

	info, err := global.NeteaseMusicSongInfo(songId)
	if err != nil {
		return nil, err
	}
	if !info.Exists() {
		return nil, errors.New("song not found")
	}
	name := info.Get("name").Str
	jumpURL := "https://y.music.163.com/m/song/" + songId
	musicURL := "http://music.163.com/song/media/outer/url?id=" + songId
	picURL := info.Get("album.picUrl").Str
	artistName := ""
	if info.Get("artists.0").Exists() {
		artistName = info.Get("artists.0.name").Str
	}
	return &message.MusicShareElement{
		MusicType:  message.CloudMusic,
		Title:      name,
		Summary:    artistName,
		Url:        jumpURL,
		PictureUrl: picURL,
		MusicUrl:   musicURL,
	}, nil
}

// based on https://github.com/wdvxdr1123/ZeroBot/blob/main/example/music/data.go
func (r *QQRobot) queryNeteaseMusic(musicName string) int64 {
	req, err := http.NewRequest("GET", "http://music.163.com/api/search/get?type=1&s="+url.QueryEscape(musicName), nil)
	if err != nil {
		return 0
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66")
	res, err := r.HttpClient.Do(req)
	if err != nil {
		return 0
	}
	data, err := ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		return 0
	}
	return gjson.ParseBytes(data).Get("result.songs.0.id").Int()
}
