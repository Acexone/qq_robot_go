package qq_robot

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
)

// 2021/10/02 5:21 by fzls

func (r *QQRobot) checkUpdates() {
	for _, rule := range r.Config.NotifyUpdate.Rules {
		lastVersion := r.CheckUpdateVersionMap[rule.Name]
		latestVersion, updateMessage := r.getLatestGitVersion(rule.GitChangelogPage)
		if version_less(lastVersion, latestVersion) {
			// 版本有更新
			r.CheckUpdateVersionMap[rule.Name] = latestVersion

			// 排除因网络连接不好而未能在启动时正确获取版本号的情况
			if lastVersion == VersionNone {
				continue
			}
			replies := r.makeNotifyUpdatesReplies(rule, latestVersion, updateMessage)
			nowStr := r.currentTime()
			for _, groupId := range rule.NotifyGroups {
				rspId := r.cqBot.SendGroupMessage(groupId, replies)
				if rspId == -1 {
					logger.Errorf("【%v Failed】 %v groupId=%v replies=%v err=%v", rule.Name, nowStr, groupId, replies, rspId)
					continue
				}
				logger.Infof("【%v】 %v groupId=%v replies=%v", rule.Name, nowStr, groupId, replies)
			}
			logger.Infof("check update %v, from %v to %v", rule.Name, lastVersion, latestVersion)
		}
	}
}

func (r *QQRobot) manualTriggerUpdateMessage(groupId int64) (replies *message.SendingMessage) {
	for _, rule := range r.Config.NotifyUpdate.Rules {
		inRange := false
		for _, group := range rule.NotifyGroups {
			if groupId == group {
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
	msg := strings.ReplaceAll(rule.Message, TemplateArgs_GitVersion, latestVersion)
	msg = strings.ReplaceAll(msg, TemplateArgs_UpdateMessage, updateMessage)
	replies.Append(message.NewText(msg))
	// 如配置了图片url，则额外发送图片
	if rule.ImageUrl != "" {
		r.tryAppendImageByUrl(replies, rule.ImageUrl)
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

var VersionNone = "v0.0.0"

// github的镜像站
var GITHUB_MIRROR_SITES = []string{
	"github.com.cnpmjs.org",
	"hub.fastgit.org",
}

func (r *QQRobot) getLatestGitVersion(gitChangelogPage string) (latestVersion string, updateMessage string) {
	var urls []string
	// 先尝试国内镜像，最后尝试直接访问
	for _, mirrorSite := range GITHUB_MIRROR_SITES {
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
	resp, err := r.HttpClient.Get(gitChangelogPage)
	if err != nil {
		logger.Errorf("getLatestGitVersion gitChangelogPage=%v err=%v", gitChangelogPage, err)
		return VersionNone, ""
	}
	defer resp.Body.Close()

	// 获取网页内容
	bytesData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("getLatestGitVersion gitChangelogPage=%v err=%v", gitChangelogPage, err)
		return VersionNone, ""
	}

	htmlText := string(bytesData)

	// 解析版本信息
	matches := regGitVersion.FindAllStringSubmatch(htmlText, -1)
	if len(matches) == 0 {
		logger.Errorf("getLatestGitVersion gitChangelogPage=%v can not find any match, html text=%v", gitChangelogPage, err, htmlText)
		return VersionNone, ""
	}

	var versions []string
	for _, match := range matches {
		versions = append(versions, match[1])
	}

	sort.Slice(versions, func(i, j int) bool {
		return !version_less(versions[i], versions[j])
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

func version_less(version_left, version_right string) bool {
	return version_less_int_list(version_to_version_int_list(version_left), version_to_version_int_list(version_right))
}

func version_less_int_list(version_int_list_left, version_int_list_right []int64) bool {
	length := len(version_int_list_left)
	if len(version_int_list_right) < length {
		length = len(version_int_list_right)
	}
	for idx := 0; idx < length; idx++ {
		if version_int_list_left[idx] != version_int_list_right[idx] {
			return version_int_list_left[idx] < version_int_list_right[idx]
		}
	}

	return len(version_int_list_left) < len(version_int_list_right)
}

// v3.2.2 => [3, 2, 2], 3.2.2 => [3, 2, 2]
func version_to_version_int_list(version string) []int64 {
	// 移除v
	if version[0] == 'v' {
		version = version[1:]
	}
	var versionIntList []int64
	for _, subVersion := range strings.Split(version, ".") {
		subVersionInt, _ := strconv.ParseInt(subVersion, 10, 64)
		versionIntList = append(versionIntList, subVersionInt)
	}
	return versionIntList
}
