package qqrobot

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Mrs4s/MiraiGo/message"
	"github.com/gookit/color"
	logger "github.com/sirupsen/logrus"
)

// 2021/10/02 5:21 by fzls

func (r *QQRobot) checkUpdates() {
	for _, rule := range r.Config.NotifyUpdate.Rules {
		lastVersion := r.CheckUpdateVersionMap[rule.Name]
		latestVersion, updateMessage := r.getLatestGitVersion(rule.GitChangelogPage)
		if versionLess(lastVersion, latestVersion) {
			// 版本有更新
			r.CheckUpdateVersionMap[rule.Name] = latestVersion

			// 排除因网络连接不好而未能在启动时正确获取版本号的情况
			if lastVersion == VersionNone {
				continue
			}
			replies := r.makeNotifyUpdatesReplies(rule, latestVersion, updateMessage)
			nowStr := r.currentTime()
			for _, groupID := range rule.NotifyGroups {
				rspID := r.cqBot.SendGroupMessage(groupID, replies)
				if rspID == -1 {
					logger.Errorf("【%v Failed】 %v groupID=%v replies=%v err=%v", rule.Name, nowStr, groupID, replies, rspID)
					continue
				}
				logger.Infof("【%v】 %v groupID=%v replies=%v", rule.Name, nowStr, groupID, replies)
			}
			logger.Infof("check update %v, from %v to %v", rule.Name, lastVersion, latestVersion)
		}
	}
}

func (r *QQRobot) manualTriggerUpdateMessage(groupID int64) (replies *message.SendingMessage) {
	for _, rule := range r.Config.NotifyUpdate.Rules {
		inRange := false
		for _, group := range rule.NotifyGroups {
			if groupID == group {
				inRange = true
				break
			}
		}
		if !inRange {
			continue
		}

		latestVersion, updateMessage := r.getLatestGitVersion(rule.GitChangelogPage)

		replies = r.makeNotifyUpdatesReplies(rule, latestVersion, updateMessage)
		logger.Infof("manualTriggerUpdateMessage %v, version=%v", rule.Name, latestVersion)

		return replies
	}

	return nil
}

func (r *QQRobot) makeNotifyUpdatesReplies(rule NotifyUpdateRule, latestVersion string, updateMessage string) *message.SendingMessage {
	// 通知相关群组
	replies := message.NewSendingMessage()

	// 在消息开头处理需要@的人
	if rule.AtAllOnTrigger {
		replies.Append(message.AtAll())
	}
	for _, atOnTrigger := range rule.AtQQsOnTrigger {
		replies.Append(message.NewAt(atOnTrigger))
	}
	msg := strings.ReplaceAll(rule.Message, templateargsGitversion, latestVersion)
	msg = strings.ReplaceAll(msg, templateargsUpdatemessage, updateMessage)
	replies.Append(message.NewText(msg))
	// 如配置了图片url，则额外发送图片
	if rule.ImageURL != "" {
		r.tryAppendImageByURL(replies, rule.ImageURL)
	}
	return replies
}

func (r *QQRobot) initCheckUpdateVersionMap() {
	if len(r.Config.NotifyUpdate.Rules) == 0 {
		return
	}
	logger.Infof(bold(color.Yellow).Render(fmt.Sprintf("将以%v的间隔定期检查配置的项目的版本更新情况", time.Second*time.Duration(r.Config.NotifyUpdate.CheckInterval))))
	for _, rule := range r.Config.NotifyUpdate.Rules {
		latestVersion, updateMessage := r.getLatestGitVersion(rule.GitChangelogPage)
		r.CheckUpdateVersionMap[rule.Name] = latestVersion
		logger.Infof(bold(color.Yellow).Render(fmt.Sprintf("项目[%v]当前的最新版本为%v, 更新信息如下：\n%v", rule.Name, latestVersion, updateMessage)))
	}
}

var regGitVersion = regexp.MustCompile(`([vV][0-9.]+)(\s+\d+\.\d+\.\d+)`)
var regUpdateInfo = regexp.MustCompile(`(更新公告</h1>)\s*<ol>((\s|\S)+?)</ol>`)
var regUpdateMessages = regexp.MustCompile("<li>(.+?)</li>")

// VersionNone 默认版本号
var VersionNone = "v0.0.0"

// GithubMirrorSites github的镜像站
var GithubMirrorSites = []string{
	"hub.fastgit.org",
	"github.com.cnpmjs.org",
}

func (r *QQRobot) getLatestGitVersion(gitChangelogPage string) (latestVersion string, updateMessage string) {
	urls := make([]string, 0, len(GithubMirrorSites)+1)
	// 先尝试国内镜像，最后尝试直接访问
	for _, mirrorSite := range GithubMirrorSites {
		urls = append(urls, strings.ReplaceAll(gitChangelogPage, "github.com", mirrorSite))
	}
	urls = append(urls, gitChangelogPage)

	for _, url := range urls {
		latestVersion, updateMessage = r._getLatestGitVersion(url)
		if latestVersion != VersionNone {
			return
		}
	}

	return
}

func (r *QQRobot) _getLatestGitVersion(gitChangelogPage string) (string, string) {
	resp, err := r.httpClient.Get(gitChangelogPage)
	if err != nil {
		logger.Debugf("getLatestGitVersion gitChangelogPage=%v err=%v", gitChangelogPage, err)
		return VersionNone, ""
	}
	defer resp.Body.Close()

	// 获取网页内容
	bytesData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Debugf("getLatestGitVersion gitChangelogPage=%v err=%v", gitChangelogPage, err)
		return VersionNone, ""
	}

	htmlText := string(bytesData)

	// 解析版本信息
	matches := regGitVersion.FindAllStringSubmatch(htmlText, -1)
	if len(matches) == 0 {
		logger.Debugf("getLatestGitVersion gitChangelogPage=%v can not find any match, err=%v, html text=%v", gitChangelogPage, err, htmlText)
		return VersionNone, ""
	}

	versions := make([]string, 0, len(matches))
	for _, match := range matches {
		versions = append(versions, match[1])
	}

	sort.Slice(versions, func(i, j int) bool {
		return !versionLess(versions[i], versions[j])
	})

	// 解析更新信息
	var updateMessage string
	if match := regUpdateInfo.FindStringSubmatch(htmlText); match != nil {
		messagesText := match[2]
		if matches := regUpdateMessages.FindAllStringSubmatch(messagesText, -1); matches != nil {
			var messages []string
			for idx, match := range matches {
				messages = append(messages, fmt.Sprintf("%v. %v", idx+1, match[1]))
			}
			updateMessage = strings.Join(messages, "\n")
		}
	}

	return versions[0], updateMessage
}

func versionLess(versionLeft, versionRight string) bool {
	return versionLessIntList(versionToVersionIntList(versionLeft), versionToVersionIntList(versionRight))
}

func versionLessIntList(versionIntListLeft, versionIntListRight []int64) bool {
	length := len(versionIntListLeft)
	if len(versionIntListRight) < length {
		length = len(versionIntListRight)
	}
	for idx := 0; idx < length; idx++ {
		if versionIntListLeft[idx] != versionIntListRight[idx] {
			return versionIntListLeft[idx] < versionIntListRight[idx]
		}
	}

	return len(versionIntListLeft) < len(versionIntListRight)
}

// v3.2.2 => [3, 2, 2], 3.2.2 => [3, 2, 2]
func versionToVersionIntList(version string) []int64 {
	// 移除v
	if version[0] == 'v' {
		version = version[1:]
	}

	subVersionList := strings.Split(version, ".")
	versionIntList := make([]int64, 0, len(subVersionList))
	for _, subVersion := range subVersionList {
		subVersionInt, _ := strconv.ParseInt(subVersion, 10, 64)
		versionIntList = append(versionIntList, subVersionInt)
	}
	return versionIntList
}
