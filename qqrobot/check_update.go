package qqrobot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/global"
	"github.com/gookit/color"
	logger "github.com/sirupsen/logrus"
)

// 2021/10/02 5:21 by fzls

type PythonDownloadNewVersionResult struct {
	Filepath string `json:"downloaded_path"`
}

func (r *QQRobot) checkUpdates() {
	for _, rule := range r.Config.NotifyUpdate.Rules {
		lastVersion := r.CheckUpdateVersionMap[rule.Name]
		latestVersion, updateMessage := r.getLatestGitVersion(rule.GitChangelogRawUrl)
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
				// 广播消息间强行间隔一秒
				time.Sleep(time.Second)
				if rspID == -1 {
					logger.Errorf("【%v Failed】 %v groupID=%v replies=%v err=%v", rule.Name, nowStr, groupID, replies, rspID)
					continue
				}
				logger.Infof("【%v】 %v groupID=%v replies=%v", rule.Name, nowStr, groupID, replies)
			}
			logger.Infof("check update %v, from %v to %v", rule.Name, lastVersion, latestVersion)

			r.updateNewVersionInGroup(rule.Name, rule.NotifyGroups, rule.DownloadNewVersionPythonInterpreterPath, rule.DownloadNewVersionPythonScriptPath)
		}
	}
}

func (r *QQRobot) updateNewVersionInGroup(ctx string, groups []int64, interpreter string, script string) {
	if interpreter != "" && script != "" && global.PathExists(interpreter) && global.PathExists(script) {
		logger.Infof("开始更新新版本到各个群中: %v", groups)
		oldVersionKeywords := "DNF蚊子腿小助手_v"

		logger.Infof("开始调用配置的更新命令来获取新版本: %v %v", interpreter, script)
		newVersionFilePath, err := downloadNewVersionUsingPythonScript(interpreter, script)
		if err != nil {
			logger.Warnf("下载新版本失败, err=%v", err)
			return
		}

		uploadFileName := filepath.Base(newVersionFilePath)

		for _, groupID := range groups {
			logger.Infof("开始上传 %v 到 群 %v", uploadFileName, groupID)
			r.updateFileInGroup(groupID, newVersionFilePath, uploadFileName, oldVersionKeywords)
			// 广播消息间强行间隔一秒
			time.Sleep(time.Second)
		}
	} else {
		logger.Infof("%v: 未配置更新python脚本，或者对应脚本不存在，将不会尝试下载并上传新版本到群文件", ctx)
	}
}

func downloadNewVersionUsingPythonScript(pythonInterpreterPath string, pythonScriptPath string) (string, error) {
	cmd := exec.Command(pythonInterpreterPath, pythonScriptPath)
	cmd.Dir = filepath.Dir(pythonScriptPath)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("调用python脚本 %v 下载新版本失败，err=%v", pythonScriptPath, err)
	}

	var result PythonDownloadNewVersionResult
	err = json.Unmarshal(out, &result)
	if err != nil {
		return "", fmt.Errorf("解析python返回的结果失败, out=%v, err=%v", string(out), err)
	}

	return result.Filepath, nil
}

func (r *QQRobot) updateFileInGroup(groupID int64, localFilePath string, uploadFileName string, oldVersionKeyWords string) {
	logger.Infof("开始更新 群 %v 的 %v 文件，旧版本关键词为 %v，新版本路径为 %v", groupID, uploadFileName, oldVersionKeyWords, localFilePath)

	// 获取群文件信息
	fs, err := r.cqBot.Client.GetGroupFileSystem(groupID)
	if err != nil {
		logger.Warnf("获取群 %v 文件系统信息失败: %v", groupID, err)
		return
	}
	files, _, err := fs.Root()
	if err != nil {
		logger.Warnf("获取群 %v 根目录文件失败: %v", groupID, err)
		return
	}

	// 移除之前版本
	for _, file := range files {
		if strings.Contains(file.FileName, oldVersionKeyWords) {
			logger.Infof("找到了目标文件=%v", file)

			res := fs.DeleteFile("", file.FileId, file.BusId)
			logger.Infof("删除群 %v 文件 %v(%v) 结果为 %v", groupID, file.FileName, file.FileId, res)
		}
	}

	// 上传新版本
	err = fs.UploadFile(localFilePath, uploadFileName, "/")
	logger.Warnf("上传群 %v 文件 %v 结果为 %v", groupID, uploadFileName, err)
}

func (r *QQRobot) manualTriggerUpdateNotify(triggerRule *Rule) (replies *message.SendingMessage) {
	for _, rule := range r.Config.NotifyUpdate.Rules {
		if rule.Name != triggerRule.Config.TargetUpdateRuleName {
			continue
		}

		latestVersion, updateMessage := r.getLatestGitVersion(rule.GitChangelogRawUrl)
		if latestVersion == VersionNone {
			break
		}

		updateMessages := r.makeNotifyUpdatesReplies(rule, latestVersion, updateMessage)
		logger.Infof("manualTriggerUpdateNotify %v, version=%v", rule.Name, latestVersion)

		// 发送给配置的目标群组
		nowStr := r.currentTime()
		for _, groupID := range rule.NotifyGroups {
			rspID := r.cqBot.SendGroupMessage(groupID, updateMessages)
			// 广播消息间强行间隔一秒
			time.Sleep(time.Second)
			if rspID == -1 {
				logger.Errorf("【%v Failed】 %v groupID=%v updateMessages=%v err=%v", rule.Name, nowStr, groupID, updateMessages, rspID)
				continue
			}
			logger.Infof("【%v】 %v groupID=%v updateMessages=%v", rule.Name, nowStr, groupID, updateMessages)
		}

		replies = message.NewSendingMessage()
		replies.Append(message.NewText("已发送更新公告到更新规则中指定的群组"))
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
		latestVersion, updateMessage := r.getLatestGitVersion(rule.GitChangelogRawUrl)
		r.CheckUpdateVersionMap[rule.Name] = latestVersion
		logger.Infof(bold(color.Yellow).Render(fmt.Sprintf("项目[%v]当前的最新版本为%v, 更新信息如下：\n%v", rule.Name, latestVersion, updateMessage)))
	}
}

var regGitVersion = regexp.MustCompile(`([vV][0-9.]+)(\s+\d+\.\d+\.\d+)`)
var regUpdateInfo = regexp.MustCompile(`更新公告\s*(?P<update_message>(\s|\S)+?)\n\n`)

// VersionNone 默认版本号
var VersionNone = "v0.0.0"

func (r *QQRobot) getLatestGitVersion(gitChangelogRawUrl string) (latestVersion string, updateMessage string) {
	urls := generateMirrorGithubRawUrls(gitChangelogRawUrl)

	for _, url := range urls {
		latestVersion, updateMessage = r._getLatestGitVersion(url)
		if latestVersion != VersionNone {
			return
		}
	}

	return
}

// 形如 https://github.com/fzls/djc_helper/raw/master/CHANGELOG.MD
var regRawUrl = regexp.MustCompile(`https://github.com/(?P<owner>\w+)/(?P<repo_name>\w+)/raw/(?P<branch_name>\w+)/(?P<filepath_in_repo>[\w\W]+)`)

func generateMirrorGithubRawUrls(gitChangelogRawUrl string) []string {
	match := regRawUrl.FindStringSubmatch(gitChangelogRawUrl)
	if match == nil {
		return []string{gitChangelogRawUrl}
	}
	owner := match[regRawUrl.SubexpIndex("owner")]
	repoName := match[regRawUrl.SubexpIndex("repo_name")]
	branchName := match[regRawUrl.SubexpIndex("branch_name")]
	filepathInRepo := match[regRawUrl.SubexpIndex("filepath_in_repo")]

	var urls []string

	// 先加入比较快的几个镜像
	urls = append(urls, "https://raw.iqiq.io/{owner}/{repo_name}/{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://raw.連接.台灣/{owner}/{repo_name}/{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://raw-gh.gcdn.mirr.one/{owner}/{repo_name}/{branch_name}/{filepath_in_repo}")

	// 随机乱序，确保均匀分布请求
	rand.Shuffle(len(urls), func(i, j int) {
		urls[i], urls[j] = urls[j], urls[i]
	})

	// 然后加入几个慢的镜像和源站
	urls = append(urls, "https://cdn.staticaly.com/gh/{owner}/{repo_name}/{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://gcore.jsdelivr.net/gh/{owner}/{repo_name}@{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://fastly.jsdelivr.net/gh/{owner}/{repo_name}@{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://raw.fastgit.org/{owner}/{repo_name}/{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://ghproxy.com/https://raw.githubusercontent.com/{owner}/{repo_name}/{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://ghproxy.futils.com/https://github.com/{owner}/{repo_name}/blob/{branch_name}/{filepath_in_repo}")

	// 最后加入原始地址和一些不可达的
	urls = append(urls, "https://raw.githubusercontents.com/{owner}/{repo_name}/{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://github.com/{owner}/{repo_name}/raw/{branch_name}/{filepath_in_repo}")

	// 替换占位符为实际值
	placeholderToValue := map[string]string{
		"{owner}":            owner,
		"{repo_name}":        repoName,
		"{branch_name}":      branchName,
		"{filepath_in_repo}": filepathInRepo,
	}
	for idx := 0; idx < len(urls); idx++ {
		for placeholder, value := range placeholderToValue {
			urls[idx] = strings.ReplaceAll(urls[idx], placeholder, value)
		}
	}

	// 返回最终结果
	return urls
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
		updateMessage = match[regUpdateInfo.SubexpIndex("update_message")]
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
