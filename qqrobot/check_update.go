package qqrobot

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Mrs4s/MiraiGo/message"
	"github.com/gookit/color"
	"github.com/pkg/errors"
	logger "github.com/sirupsen/logrus"

	"github.com/Mrs4s/go-cqhttp/global"
)

// 2021/10/02 5:21 by fzls

// PythonDownloadNewVersionResult 使用python脚本下载最新版本的结果
type PythonDownloadNewVersionResult struct {
	Filepath string `json:"downloaded_path"`
}

func (r *QQRobot) checkUpdates() {
	for _, rule := range r.Config.NotifyUpdate.Rules {
		lastVersion := r.CheckUpdateVersionMap[rule.Name]
		latestVersion, updateMessage := r.getLatestGitVersion(rule.GitChangelogRawURL)
		if versionLess(lastVersion, latestVersion) {
			// 版本有更新
			r.CheckUpdateVersionMap[rule.Name] = latestVersion

			// 排除因网络连接不好而未能在启动时正确获取版本号的情况
			if lastVersion == VersionNone {
				continue
			}

			// re: 以下是临时措施，应对changelog中的版本更新后平均约 10 分钟后github action的打包release流程才完成的情况
			// undone: 等将新版本改成基于release中的meta信息来获取后，再改成实时通知
			go func() {
				logger.Warnf("%v 版本有更新 %v => %v，但目前因获取版本的来源比release早约十分钟，因此在这里等待20分钟后再实际进行通知", rule.Name, lastVersion, latestVersion)
				select {
				case <-time.After(2 * 10 * time.Minute):
					break
				case <-r.quitCtx.Done():
					return
				}

				logger.Infof("%v 版本有更新 %v => %v, 开始通知各个群以及上传群文件", rule.Name, lastVersion, latestVersion)

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

				r.updateNewVersionInGroup(rule.Name, rule.NotifyGroups, rule.DownloadNewVersionPythonInterpreterPath, rule.DownloadNewVersionPythonScriptPath, true)
			}()
		}
	}
}

func (r *QQRobot) updateNewVersionInGroup(ctx string, groups []int64, interpreter string, script string, needRetry bool) {
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

		groupsToUpload := groups
		failIndex := 1
		retryWaitTime := time.Minute
		maxRetryWaitTime := time.Hour
		for {
			// 尝试上传新版本
			for _, groupID := range groupsToUpload {
				logger.Infof("开始上传 %v 到 群 %v", uploadFileName, groupID)
				r.updateFileInGroup(groupID, newVersionFilePath, uploadFileName, oldVersionKeywords, false)
				// 广播消息间强行间隔一秒
				time.Sleep(time.Second)
			}

			// 检查是否上传成功
			var groupsNotUploaded []int64
			for _, groupID := range groupsToUpload {
				if !r.hasFileInGroup(groupID, uploadFileName) {
					groupsNotUploaded = append(groupsNotUploaded, groupID)
				}
			}
			if len(groupsNotUploaded) == 0 {
				logger.Infof("全部上传成功，完毕")
				break
			}

			logger.Infof("%v 个群未上传成功： %v", len(groupsNotUploaded), groupsNotUploaded)
			if !needRetry {
				logger.Warnf("当前配置为不需要重试")
				break
			}

			logger.Infof("第 %v 次上传失败， 等待 %v 后再尝试上传到这些群中", failIndex, retryWaitTime)
			select {
			case <-time.After(retryWaitTime):
				break
			case <-r.quitCtx.Done():
				return
			}

			groupsToUpload = groupsNotUploaded
			failIndex += 1
			retryWaitTime *= 2
			if retryWaitTime >= maxRetryWaitTime {
				retryWaitTime = maxRetryWaitTime
			}
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
		return "", errors.Errorf("调用python脚本 %v 下载新版本失败，err=%v, out=%v", pythonScriptPath, err, string(out))
	}

	// 现在脚本基于github来下载，中间需要二次压缩，导致会出现一些多余的字符串，因此这里需要将结果部分提取出来
	output := string(out)
	boundaryMark := "$$boundary$$"
	parts := strings.Split(output, boundaryMark)
	if len(parts) != 3 {
		return "", errors.Errorf("输出格式不符合预期，预期应由 %v 分隔为三部分, output=%v", boundaryMark, output)
	}

	jsonResult := parts[1]
	jsonResult = strings.TrimSpace(jsonResult)

	var result PythonDownloadNewVersionResult
	err = json.Unmarshal([]byte(jsonResult), &result)
	if err != nil {
		return "", errors.Errorf("解析python返回的结果失败, jsonResult=%v, err=%v", jsonResult, err)
	}

	return result.Filepath, nil
}

func (r *QQRobot) updateFileInGroup(
	groupID int64,
	localFilePath string,
	uploadFileName string,
	oldVersionKeyWords string,
	replaceIfExists bool,
) {
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

	if !replaceIfExists {
		// 如果已经存在该文件，则直接返回
		for _, file := range files {
			if file.FileName == uploadFileName {
				logger.Infof("群 %v 中已有 %v 文件，且当前配置为不覆盖已有文件，将直接跳过", groupID, uploadFileName)
				return
			}
		}
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

// hasFileInGroup 指定群中是否有指定名称的文件
func (r *QQRobot) hasFileInGroup(groupID int64, uploadFileName string) (has bool) {
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
		if file.FileName == uploadFileName {
			return true
		}
	}

	return
}

func (r *QQRobot) manualTriggerUpdateNotify(triggerRule *Rule) (replies *message.SendingMessage) {
	for _, rule := range r.Config.NotifyUpdate.Rules {
		if rule.Name != triggerRule.Config.TargetUpdateRuleName {
			continue
		}

		latestVersion, updateMessage := r.getLatestGitVersion(rule.GitChangelogRawURL)
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
		latestVersion, updateMessage := r.getLatestGitVersion(rule.GitChangelogRawURL)
		r.CheckUpdateVersionMap[rule.Name] = latestVersion
		logger.Infof(bold(color.Yellow).Render(fmt.Sprintf("项目[%v]当前的最新版本为%v, 更新信息如下：\n%v", rule.Name, latestVersion, updateMessage)))
	}
}

var regGitVersion = regexp.MustCompile(`([vV][0-9.]+)(\s+\d+\.\d+\.\d+)`)
var regUpdateInfo = regexp.MustCompile(`更新公告\s*(?P<update_message>(\s|\S)+?)\n\n`)

// VersionNone 默认版本号
var VersionNone = "v0.0.0"

func (r *QQRobot) getLatestGitVersion(gitChangelogRawURL string) (latestVersion string, updateMessage string) {
	urls := generateMirrorGithubRawUrls(gitChangelogRawURL)

	for _, url := range urls {
		latestVersion, updateMessage = r._getLatestGitVersion(url)
		if latestVersion != VersionNone {
			return
		}
	}

	return
}

// 形如 https://github.com/fzls/djc_helper/raw/master/CHANGELOG.MD
var regRawURL = regexp.MustCompile(`https://github.com/(?P<owner>\w+)/(?P<repo_name>\w+)/raw/(?P<branch_name>\w+)/(?P<filepath_in_repo>[\w\W]+)`)

func generateMirrorGithubRawUrls(gitChangelogRawURL string) []string {
	match := regRawURL.FindStringSubmatch(gitChangelogRawURL)
	if match == nil {
		return []string{gitChangelogRawURL}
	}
	owner := match[regRawURL.SubexpIndex("owner")]
	repoName := match[regRawURL.SubexpIndex("repo_name")]
	branchName := match[regRawURL.SubexpIndex("branch_name")]
	filepathInRepo := match[regRawURL.SubexpIndex("filepath_in_repo")]

	var urls []string

	// 先加入比较快的几个镜像
	urls = append(urls, "https://hk1.monika.love/{owner}/{repo_name}/{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://raw.iqiq.io/{owner}/{repo_name}/{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://raw-gh.gcdn.mirr.one/{owner}/{repo_name}/{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://raw.fastgit.org/{owner}/{repo_name}/{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://raw.githubusercontents.com/{owner}/{repo_name}/{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://gcore.jsdelivr.net/gh/{owner}/{repo_name}@{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://kgithub.com/{owner}/{repo_name}/raw/{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://cdn.staticaly.com/gh/{owner}/{repo_name}/{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://ghproxy.com/https://raw.githubusercontent.com/{owner}/{repo_name}/{branch_name}/{filepath_in_repo}")

	// 随机乱序，确保均匀分布请求
	rand.Shuffle(len(urls), func(i, j int) {
		urls[i], urls[j] = urls[j], urls[i]
	})

	// 然后加入几个慢的镜像和源站
	urls = append(urls, "https://fastly.jsdelivr.net/gh/{owner}/{repo_name}@{branch_name}/{filepath_in_repo}")
	urls = append(urls, "https://cdn.jsdelivr.net/gh/{owner}/{repo_name}@{branch_name}/{filepath_in_repo}")

	// 最后加入原始地址和一些不可达的
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
	bytesData, err := io.ReadAll(resp.Body)
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
