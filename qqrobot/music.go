package qqrobot

import (
	"io"
	"net/http"
	"net/url"

	"github.com/Mrs4s/MiraiGo/message"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/Mrs4s/go-cqhttp/global"
)

func (r *QQRobot) makeMusicShareElement(musicName string, musicType int) (*message.MusicShareElement, error) {
	switch musicType {
	case message.CloudMusic:
		return r.makeCloudMusicShareElement(musicName)
	case message.QQMusic:
		return r.makeQQMusicShareElement(musicName)
	default:
		return nil, errors.Errorf("未知音乐类型=%v", musicType)
	}
}

// 基于 https://github.com/wdvxdr1123/ZeroBot/blob/main/example/music/data.go
func (r *QQRobot) makeCloudMusicShareElement(musicName string) (*message.MusicShareElement, error) {
	songID := r.queryNeteaseMusic(musicName)

	info, err := global.NeteaseMusicSongInfo(songID)
	if err != nil {
		return nil, err
	}
	if !info.Exists() {
		return nil, errors.New("song not found")
	}
	name := info.Get("name").Str
	jumpURL := "https://y.music.163.com/m/song/" + songID
	musicURL := "http://music.163.com/song/media/outer/url?id=" + songID
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

func (r *QQRobot) queryNeteaseMusic(musicName string) string {
	req, err := http.NewRequest("GET", "http://music.163.com/api/search/get?type=1&s="+url.QueryEscape(musicName), nil)
	if err != nil {
		return "0"
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66")
	res, err := r.httpClient.Do(req)
	if err != nil {
		return "0"
	}
	data, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		return "0"
	}
	return gjson.ParseBytes(data).Get("result.songs.0.id").String()
}

// 基于 https://github.com/FloatTech/ZeroBot-Plugin/blob/master/plugin_music/selecter.go
func (r *QQRobot) makeQQMusicShareElement(musicName string) (*message.MusicShareElement, error) {
	songID := r.queryQQMusic(musicName)

	info, err := global.QQMusicSongInfo(songID)
	if err != nil {
		return nil, err
	}
	if !info.Get("track_info").Exists() {
		return nil, errors.New("song not found")
	}
	name := info.Get("track_info.name").Str
	mid := info.Get("track_info.mid").Str
	albumMid := info.Get("track_info.album.mid").Str
	pinfo, _ := global.GetBytes("http://u.y.qq.com/cgi-bin/musicu.fcg?g_tk=2034008533&uin=0&format=json&data={\"comm\":{\"ct\":23,\"cv\":0},\"url_mid\":{\"module\":\"vkey.GetVkeyServer\",\"method\":\"CgiGetVkey\",\"param\":{\"guid\":\"4311206557\",\"songmid\":[\"" + mid + "\"],\"songtype\":[0],\"uin\":\"0\",\"loginflag\":1,\"platform\":\"23\"}}}&_=1599039471576")
	jumpURL := "https://i.y.qq.com/v8/playsong.html?platform=11&appshare=android_qq&appversion=10030010&hosteuin=oKnlNenz7i-s7c**&songmid=" + mid + "&type=0&appsongtype=1&_wv=1&source=qq&ADTAG=qfshare"
	purl := gjson.ParseBytes(pinfo).Get("url_mid.data.midurlinfo.0.purl").Str
	preview := "http://y.gtimg.cn/music/photo_new/T002R180x180M000" + albumMid + ".jpg"
	content := info.Get("track_info.singer.0.name").Str
	return &message.MusicShareElement{
		MusicType:  message.QQMusic,
		Title:      name,
		Summary:    content,
		Url:        jumpURL,
		PictureUrl: preview,
		MusicUrl:   purl,
	}, nil
}

func (r *QQRobot) queryQQMusic(musicName string) string {
	// 搜索音乐信息 第一首歌
	h1 := http.Header{
		"User-Agent": []string{"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:84.0) Gecko/20100101 Firefox/84.0"},
	}
	search, _ := url.Parse("https://c.y.qq.com/soso/fcgi-bin/client_search_cp")
	search.RawQuery = url.Values{
		"w": []string{musicName},
	}.Encode()
	res := netGet(search.String(), h1)
	info := gjson.ParseBytes(res[9 : len(res)-1]).Get("data.song.list.0")

	return info.Get("songid").String()
}

// netGet 返回请求数据
func netGet(url string, header http.Header) []byte {
	client := &http.Client{}
	request, _ := http.NewRequest("GET", url, nil)
	request.Header = header
	res, err := client.Do(request)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	result, _ := io.ReadAll(res.Body)
	return result
}
