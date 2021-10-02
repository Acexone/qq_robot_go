package qq_robot

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/coolq"
	"github.com/gookit/color"
)

type MessageKey struct {
	MessageId int64
	SenderQQ  int64
	GroupId   int64
}

func makeMessageKey(messageId, senderQQ, groupId int64) MessageKey {
	return MessageKey{
		MessageId: messageId,
		SenderQQ:  senderQQ,
		GroupId:   groupId,
	}
}

type QQRobot struct {
	cqBot *coolq.CQBot

	Config    Config // 配置
	StartTime time.Time

	Rules                                        []*Rule
	RuleTypeToMessageIdToRuleApplyCount          map[RuleType]map[MessageKey]int32 // 消息类型 => 消息Key => 该消息被当前类型规则处理的次数
	GroupToMemberToTriggerRuleTimes              map[int64]map[int64][]int64       // group => qq => list of 触发规则的时间戳
	GroupToRuleNameToLastSuccessTriggerTimestamp map[int64]map[string]int64        // group => rulename => 上次在cd外成功触发的时间戳

	HttpClient http.Client

	CheckUpdateVersionMap map[string]string // 配置的检查更新名称=>最近的版本号，如"DNF蚊子腿小助手更新"=>"v4.2.2"

	quitCtx  context.Context
	quitFunc context.CancelFunc
}

func NewQQRobot(cqRobot *coolq.CQBot, configPath string) *QQRobot {
	config := LoadConfig(configPath)

	r := &QQRobot{
		cqBot: cqRobot,

		Config:                                       config,
		RuleTypeToMessageIdToRuleApplyCount:          map[RuleType]map[MessageKey]int32{},
		GroupToMemberToTriggerRuleTimes:              map[int64]map[int64][]int64{},
		GroupToRuleNameToLastSuccessTriggerTimestamp: map[int64]map[string]int64{},
		HttpClient:                                   http.Client{Timeout: time.Duration(config.Robot.Timeout) * time.Second},
		CheckUpdateVersionMap:                        map[string]string{},
	}
	for _, config := range config.Rules {
		r.Rules = append(r.Rules, NewRule(config))
	}
	return r
}
func (r *QQRobot) Start() {
	r.StartTime = time.Now()
	r.quitCtx, r.quitFunc = context.WithCancel(context.Background())

	r.notify(r.Config.Robot.OnStart)
	go r.ticker()
}

func (r *QQRobot) Stop() {
	r.notify(r.Config.Robot.OnStop)
	r.quitFunc()
}

func (r *QQRobot) notify(cfg NotifyConfig) {
	msgTemplate := cfg.Message
	if cfg.Name == "机器人下线" {
		msgTemplate = strings.ReplaceAll(msgTemplate, TemplateArgs_WorkTime, time.Since(r.StartTime).String())
	}
	if r.Config.Robot.Debug {
		logger.Debug("debug mode, do not notify", cfg.Name, cfg.NotifyGroups, msgTemplate)
		return
	}
	msg := message.NewSendingMessage()
	msg.Append(message.NewText(msgTemplate))
	nowStr := r.currentTime()
	for _, groupId := range cfg.NotifyGroups {
		if r.cqBot.Client.FindGroup(groupId) == nil {
			// 不在该群里，跳过
			continue
		}
		retCode := r.cqBot.SendGroupMessage(groupId, msg)
		if retCode == -1 {
			logger.Errorf("【%v Failed】 %v groupId=%v message=%v err=%v", cfg.Name, nowStr, groupId, msg, retCode)
			return
		}
		logger.Infof("【%v】 %v groupId=%v message=%v", cfg.Name, nowStr, groupId, msg)
	}
	logger.Infof("robot on %v finished", cfg.Name)
}

func (r *QQRobot) ticker() {
	if r.Config.NotifyUpdate.CheckInterval <= 0 {
		return
	}

	checkUpdateTicker := time.NewTicker(time.Second * time.Duration(r.Config.NotifyUpdate.CheckInterval))
	defer checkUpdateTicker.Stop()

	r.initCheckUpdateVersionMap()

	for {
		// 开始监听
		select {
		case <-checkUpdateTicker.C:
			r.checkUpdates()
		case <-r.quitCtx.Done():
			return
		}
	}
}

func (r *QQRobot) RegisterHandlers() {
	// TODO: re: 添加其他事件的处理 @2021-10-02 05:35:37

	r.cqBot.Client.OnGroupMessage(r.OnGroupMessage)
	r.cqBot.Client.OnPrivateMessage(r.OnPrivateMessage)
	//r.cqBot.Client.OnSelfPrivateMessage(rprivateMessageEvent)
	//r.cqBot.Client.OnSelfGroupMessage(rgroupMessageEvent)
	r.cqBot.Client.OnTempMessage(r.OnTempMessage)
	//r.cqBot.Client.OnGroupMuted(rgroupMutedEvent)
	//r.cqBot.Client.OnGroupMessageRecalled(rgroupRecallEvent)
	//r.cqBot.Client.OnGroupNotify(rgroupNotifyEvent)
	//r.cqBot.Client.OnFriendNotify(rfriendNotifyEvent)
	//r.cqBot.Client.OnMemberSpecialTitleUpdated(rmemberTitleUpdatedEvent)
	//r.cqBot.Client.OnFriendMessageRecalled(rfriendRecallEvent)
	//r.cqBot.Client.OnReceivedOfflineFile(rofflineFileEvent)
	//r.cqBot.Client.OnJoinGroup(rjoinGroupEvent)
	//r.cqBot.Client.OnLeaveGroup(rleaveGroupEvent)
	r.cqBot.Client.OnGroupMemberJoined(r.OnGroupMemberJoined)
	//r.cqBot.Client.OnGroupMemberLeaved(rmemberLeaveEvent)
	//r.cqBot.Client.OnGroupMemberPermissionChanged(rmemberPermissionChangedEvent)
	//r.cqBot.Client.OnGroupMemberCardUpdated(rmemberCardUpdatedEvent)
	//r.cqBot.Client.OnNewFriendRequest(rfriendRequestEvent)
	//r.cqBot.Client.OnNewFriendAdded(rfriendAddedEvent)
	//r.cqBot.Client.OnGroupInvited(rgroupInvitedEvent)
	//r.cqBot.Client.OnUserWantJoinGroup(rgroupJoinReqEvent)
	//r.cqBot.Client.OnOtherClientStatusChanged(rotherClientStatusChangedEvent)
	//r.cqBot.Client.OnGroupDigest(rgroupEssenceMsg)
}

func (r *QQRobot) OnGroupMessage(client *client.QQClient, m *message.GroupMessage) {
	for _, rule := range r.Rules {
		if err := r.applyGroupRule(m, rule); err != nil {
			return
		}
	}
}

func (r *QQRobot) OnGroupMemberJoined(client *client.QQClient, m *client.MemberJoinGroupEvent) {
	for _, rule := range r.Rules {
		if err := r.onMemberJoin(m, rule); err != nil {
			return
		}
	}
}

func (r *QQRobot) OnPrivateMessage(client *client.QQClient, m *message.PrivateMessage) {
	r.onPrivateOrTempMessage(m.Sender.Uin, 0, 0, m)
}

func (r *QQRobot) OnTempMessage(client *client.QQClient, m *client.TempMessageEvent) {
	r.onPrivateOrTempMessage(0, m.Message.GroupCode, m.Message.Sender.Uin, m)
}

func (r *QQRobot) applyGroupRule(m *message.GroupMessage, rule *Rule) error {
	config := rule.Config
	nowStr := r.currentTime()
	nowUnix := time.Now().Unix()

	groupId := m.GroupCode
	senderUin := m.Sender.Uin
	senderName := m.Sender.Nickname

	senderInfo, err := r.cqBot.Client.GetMemberInfo(groupId, senderUin)
	isAdmin := false
	if err == nil {
		isAdmin = isMemberAdmin(senderInfo.Permission)
	}

	if _, ok := config.GroupIds[groupId]; !ok {
		return nil
	}

	// 获取消息id，以及判断匹配了关键词
	var source int64
	var atGivenUsers bool
	var hitKeyWords bool
	var hitKeyWordString string
	var triggerTooOften bool

	source = int64(m.Id)

	for _, msg := range m.Elements {
		if atMsg, ok := msg.(*message.AtElement); ok {
			for _, target := range config.AtQQs {
				if atMsg.Target == target {
					atGivenUsers = true
					break
				}
			}
		}
		if msg.Type() == message.Text || msg.Type() == message.Image || msg.Type() == message.LightApp {
			if !hitKeyWords {
			OuterLoop:
				for _, keywordRegex := range config.KeywordRegexes {
					var text string
					switch msgVal := msg.(type) {
					case *message.TextElement:
						text = msgVal.Content
					case *message.LightAppElement:
						text = msgVal.Content
					case *message.GroupImageElement:
						text = fmt.Sprintf("%v\n%v", msgVal.Url, r.ocr(msgVal))
					}
					if keywordRegex.MatchString(text) {
						for _, excludeKeywordRegex := range config.ExcludeKeywordRegexes {
							if excludeKeywordRegex.MatchString(text) {
								continue OuterLoop
							}
						}

						hitKeyWords = true
						hitKeyWordString = keywordRegex.String()
						break
					}
				}
			}
		}
	}
	if config.TriggerRuleCount != 0 && config.TriggerRuleDuration != 0 && r.GroupToMemberToTriggerRuleTimes[groupId] != nil {
		// 计算是否这个QQ在滥用机器人功能
		var triggerCount int64
		checkStartTime := time.Now().Unix() - config.TriggerRuleDuration
		triggerTimes := r.GroupToMemberToTriggerRuleTimes[groupId][senderUin]
		startIdx := sort.Search(len(triggerTimes), func(i int) bool {
			return triggerTimes[i] >= checkStartTime
		})
		for idx := startIdx; idx < len(triggerTimes); idx++ {
			if triggerTimes[idx] >= checkStartTime {
				triggerCount++
			}
		}
		if triggerCount >= config.TriggerRuleCount {
			triggerTooOften = true
		}
	}

	// 是否已经回复
	maybeKilledWrongPerson := false // 误杀
	if _, replied := rule.ProcessedMessages[source]; replied {
		maybeKilledWrongPerson = true
		logger.Warnf("【似乎消息混了，不过没办法，继续处理吧-。-】", nowStr, config.Name, p(m))
	}
	// 是否不需要回复
	if len(config.KeywordRegexes) != 0 && !hitKeyWords ||
		len(config.AtQQs) != 0 && !atGivenUsers ||
		config.TriggerRuleCount != 0 && config.TriggerRuleDuration != 0 && !triggerTooOften ||
		!hitKeyWords && !atGivenUsers && !triggerTooOften {
		return nil
	}

	// 判断是否在规定的时间段内
	if len(config.TimePeriods) != 0 {
		valid := false
		now := time.Now()
		for _, tp := range config.TimePeriods {
			if now.Second() < tp.StartSecond {
				continue
			}
			if tp.EndSecond != 0 && now.Second() > tp.EndSecond {
				continue
			}
			if now.Minute() < tp.StartMinute {
				continue
			}
			if tp.EndMinute != 0 && now.Minute() > tp.EndMinute {
				continue
			}
			if now.Hour() < tp.StartHour {
				continue
			}
			if tp.EndHour != 0 && now.Hour() > tp.EndHour {
				continue
			}
			var weekday int
			if now.Weekday() == time.Sunday {
				weekday = 7
			} else {
				weekday = int(now.Weekday())
			}
			if tp.StartWeekDay != 0 && weekday < tp.StartWeekDay {
				continue
			}
			if tp.EndWeekDay != 0 && weekday > tp.EndWeekDay {
				continue
			}

			valid = true
			break
		}
		if !valid {
			return nil
		}
	}

	// 判断是否是排除的用户列表
	for _, exclude_qq := range config.ExcludeQQs {
		if exclude_qq == senderUin {
			logger.Info("【ExcludedQQ】", nowStr, config.Name, p(m))
			return nil
		}
	}
	if config.ExcludeAdmin && isAdmin {
		logger.Info("【ExcludedAdmin】", nowStr, config.Name, p(m))
		return nil
	}

	messageApplyCount := r.RuleTypeToMessageIdToRuleApplyCount[config.Type]
	if messageApplyCount == nil {
		messageApplyCount = map[MessageKey]int32{}
		r.RuleTypeToMessageIdToRuleApplyCount[config.Type] = messageApplyCount
	}
	ruleTypeConfig := RuleTypeConfig{}
	for _, cfg := range r.Config.RuleTypeConfigs {
		if cfg.Type == config.Type {
			ruleTypeConfig = cfg
			break
		}
	}
	messageKey := makeMessageKey(source, senderUin, groupId)
	if ruleTypeConfig.MaxApplyCount != RuleTypeMaxApplyCount_Infinite && messageApplyCount[messageKey] >= ruleTypeConfig.MaxApplyCount {
		return nil
	}

	guideContent := config.GuideContent
	// 判断是否在cd内触发了规则
	if config.CD != 0 && r.GroupToRuleNameToLastSuccessTriggerTimestamp[groupId] != nil {
		lastTriggerTime := r.GroupToRuleNameToLastSuccessTriggerTimestamp[groupId][config.Name]
		if nowUnix < lastTriggerTime+config.CD {
			if len(config.GuideContentInCD) == 0 {
				// 未设置cd内回复内容，则视为未触发
				logger.Info("【InCD】", nowStr, config.Name, p(m))
				return nil
			} else {
				// 替换回复内容为cd回复内容
				guideContent = strings.ReplaceAll(config.GuideContentInCD, TemplateArgs_CD, strconv.FormatInt(config.CD, 10))
			}
		}
	}

	// 记录这个QQ触发规则的时间戳
	if r.GroupToMemberToTriggerRuleTimes[groupId] == nil {
		r.GroupToMemberToTriggerRuleTimes[groupId] = map[int64][]int64{}
	}
	r.GroupToMemberToTriggerRuleTimes[groupId][senderUin] = append(r.GroupToMemberToTriggerRuleTimes[groupId][senderUin], time.Now().Unix())

	// 记录这个规则触发的时间戳
	if r.GroupToRuleNameToLastSuccessTriggerTimestamp[groupId] == nil {
		r.GroupToRuleNameToLastSuccessTriggerTimestamp[groupId] = map[string]int64{}
	}
	r.GroupToRuleNameToLastSuccessTriggerTimestamp[groupId][config.Name] = nowUnix

	// ok

	// 回复消息

	muteTime := config.MuteTime
	if config.ParseMuteTime {
		// 解析消息内容，判定禁言时间
		muteTime = parseMuteTime(m.Elements)
		if muteTime == 0 {
			// 其实是不符合禁言套餐规则，直接返回
			return nil
		}
	}
	muteTime = truncatMuteTime(muteTime)

	// 自动回复关键词
	replies := message.NewSendingMessage()

	// 在消息开头处理需要@的人
	if config.AtAllOnTrigger {
		replies.Append(message.AtAll())
	}
	for _, atOnTrigger := range config.AtQQsOnTrigger {
		replies.Append(message.NewAt(atOnTrigger))
	}

	switch config.Action {
	case ActionType_Guide:
		guideContent = strings.ReplaceAll(guideContent, TemplateArgs_MuteTime, strconv.FormatInt(muteTime, 10))
		if config.GitChangelogPage != "" {
			latestVersion, updateMessage := r.getLatestGitVersion(config.GitChangelogPage)
			guideContent = strings.ReplaceAll(guideContent, TemplateArgs_GitVersion, latestVersion)
			guideContent = strings.ReplaceAll(guideContent, TemplateArgs_UpdateMessage, updateMessage)
		}
		if len(guideContent) != 0 {
			replies.Append(message.NewText(guideContent))
		}
	case ActionType_Command:
		for _, msg := range m.Elements {
			msgVal, ok := msg.(*message.TextElement)
			if !ok {
				continue
			}

			for _, keywordRegex := range config.KeywordRegexes {
				if keywordRegex.MatchString(msgVal.Content) {
					commandStr := keywordRegex.ReplaceAllString(msgVal.Content, "")
					err, msg, extraReplies := r.processCommand(commandStr, m)
					if err != nil {
						replies.Append(message.NewText(guideContent))
						replies.Elements = append(replies.Elements, extraReplies...)
						errMsg := err.Error()
						if e, ok := err.(*exec.ExitError); ok {
							errMsg += fmt.Sprintf("\n%v", string(e.Stderr))
						}
						logger.Errorf("执行指令出错, rule=%v err=%v", config.Name, errMsg)
						continue
					}
					replies.Append(message.NewText(fmt.Sprintf("指令【%v】执行成功\n结果为【%v】\n", commandStr, msg)))
					replies.Elements = append(replies.Elements, extraReplies...)
				}
			}
		}
	case ActionType_Food:
		extraReplies, err := r.createFoodMessage(rule)
		replies.Elements = append(replies.Elements, extraReplies.Elements...)
		if err != nil {
			logger.Errorf("createFoodMessage, rule=%v err=%v", config.Name, err)
			replies.Append(message.NewText(guideContent))
		}
	case ActionType_AiChat:
		var chatText string
		for _, msg := range m.Elements {
			if msgVal, ok := msg.(*message.TextElement); ok {
				chatText += msgVal.Content
			}
		}
		reply := r.aiChat(senderUin, chatText)
		// 特殊替换一下
		if notFoundMessage := r.Config.Robot.ChatAnswerNotFoundMessage; len(notFoundMessage) != 0 {
			reply = strings.ReplaceAll(reply, "chat answer not found", notFoundMessage)
		}
		if reply != "" {
			replies.Append(message.NewText(reply))
		}
	case ActionType_SendUpdateMessage:
		if isAdmin {
			// 手动触发更新通知
			if res := r.manualTriggerUpdateMessage(groupId); res != nil {
				replies.Elements = append(replies.Elements, res.Elements...)
			} else {
				replies.Append(message.NewText("当前群组没有配置检查更新哦~"))
			}
		} else {
			replies.Append(message.NewText("只有管理员可以执行这个指令哦~不要调皮<_<"))
		}
	case ActionType_Repeater:
		if isAdmin {
			// 复读内容到指定的群组
			// 移除首行（首行设定为关键词）
			for idx := 0; idx < len(m.Elements); idx++ {
				msg := m.Elements[idx]

				if msgVal, ok := msg.(*message.TextElement); ok {
					lines := strings.Split(msgVal.Content, "\n")
					if len(lines) > 0 {
						lines[0] = fmt.Sprintf("下面由无情的复读机为您转播来自 %v 的消息：\n------------------------------", senderName)
					}
					msgVal.Content = strings.Join(lines, "\n")
					break
				}
			}
			for idx, repeatMessages := range r.getForwardMessagesList(m, true) {
				if idx == 0 {
					// 第一条转发的消息加上 @all
					repeatMessages.Elements = append([]message.IMessageElement{message.AtAll()}, repeatMessages.Elements...)
				}

				for _, repeatToGroup := range config.RepeatToGroups {
					forwardRspId := r.cqBot.SendGroupMessage(repeatToGroup, repeatMessages)
					if forwardRspId == -1 {
						logger.Error(fmt.Sprintf("【RepeatToGroup(%v) Failed】", repeatToGroup), nowStr, config.Name, repeatMessages, forwardRspId)
						continue
					}
					logger.Info(fmt.Sprintf("【RepeatToGroup(%v)】", repeatToGroup), nowStr, config.Name, repeatMessages, forwardRspId)
				}
			}
		} else {
			replies.Append(message.NewText("只有管理员可以执行这个指令哦~不要调皮<_<"))
		}
	default:
		if hitKeyWords {
			replies.Append(message.NewText(fmt.Sprintf(
				"提问前请先看群文件中【常见问题解答】与【手动安装运行环境教程】文档，如果看完仍旧不能解疑，欢迎提问。\n" +
					"但是如果是文档中已回答的问题，时间有限，恕不回答.\n" +
					"来自自动回复机器人~")))
		}
	}

	// 如配置了图片url，则额外发送图片
	imageUrl := config.ImageUrl
	if len(config.RandomImageUrls) != 0 {
		randIdx := rand.Intn(len(config.RandomImageUrls))
		imageUrl = config.RandomImageUrls[randIdx]
	}
	if imageUrl != "" {
		r.tryAppendImageByUrl(replies, imageUrl)
	}

	if maybeKilledWrongPerson {
		replies.Append(message.NewText("似乎前面有人代替你被误杀了。但是，正义的铁拳虽然会乱锤，却不会错过正确的人。宁可错杀三千，不可放过一人！（手动眼部红光特效）"))
		r.tryAppendImageByUrl(replies, "https://s3.ax1x.com/2021/03/17/66NRi9.gif")
	}

	if len(replies.Elements) != 0 {
		keyWord := fmt.Sprintf("hitKeyWordString=%v", hitKeyWordString)

		// 补充reply信息
		replies.Elements = append([]message.IMessageElement{message.NewReply(m)}, replies.Elements...)

		rspId := r.cqBot.SendGroupMessage(groupId, replies)
		if rspId == -1 {
			logger.Error("【ReplyFail】", nowStr, config.Name, keyWord, p(m), rspId)
			return err
		}
		logger.Info(color.Style{color.Bold, color.Green}.Renderln("【OK】", nowStr, config.Name, keyWord, p(m), source, replies, rspId))
	}

	if config.RevokeMessage {
		err := r.cqBot.Client.RecallGroupMessage(groupId, m.Id, m.InternalId)
		if err != nil {
			logger.Error("【RevokeMessage Fail】", nowStr, config.Name, p(m), err)
		} else {
			logger.Info("【RevokeMessage OK】", nowStr, config.Name, p(m), source, replies)
		}
	}
	if muteTime != 0 {
		if senderInfo != nil {
			err := senderInfo.Mute(uint32(muteTime))
			if err != nil {
				logger.Error("【Mute Fail】", nowStr, config.Name, p(m), err)
			} else {
				logger.Info("【Mute OK】", nowStr, config.Name, p(m), source)
			}
		} else {
			logger.Info("【Mute Fail Not Found】", nowStr, config.Name, p(m), source)
		}
	}

	// 处理转发
	if needForward := len(config.ForwardToQQs) != 0 || len(config.ForwardToGroups) != 0; needForward {
		for _, forwardMessages := range r.getForwardMessagesList(m, false) {
			for _, forwardToQQ := range config.ForwardToQQs {
				forwardRspId := r.cqBot.SendPrivateMessage(forwardToQQ, 0, forwardMessages)
				if forwardRspId == -1 {
					logger.Error(fmt.Sprintf("【ForwardToQQ(%v) Failed】", forwardToQQ), nowStr, config.Name, forwardMessages, forwardRspId)
					continue
				}
				logger.Info(fmt.Sprintf("【ForwardToQQ(%v)】", forwardToQQ), nowStr, config.Name, forwardMessages, forwardRspId)
			}
			for _, forwardToGroup := range config.ForwardToGroups {
				forwardRspId := r.cqBot.SendGroupMessage(forwardToGroup, forwardMessages)
				if forwardRspId == -1 {
					logger.Error(fmt.Sprintf("【ForwardToGroup(%v) Failed】", forwardToGroup), nowStr, config.Name, forwardMessages, forwardRspId)
					continue
				}
				logger.Info(fmt.Sprintf("【ForwardToGroup(%v)】", forwardToGroup), nowStr, config.Name, forwardMessages, forwardRspId)
			}
		}
	}

	messageApplyCount[messageKey]++
	rule.ProcessedMessages[source] = struct{}{}

	return nil
}

func (r *QQRobot) getForwardMessagesList(m *message.GroupMessage, forRepeat bool) []*message.SendingMessage {
	var forwardMessagesList []*message.SendingMessage

	messages := m.Elements

	for len(messages) != 0 {
		forwardMessages := message.NewSendingMessage()
		leftMessages := message.NewSendingMessage()

		msgSize := 0
		for _, msg := range messages {
			switch msgVal := msg.(type) {
			case *message.TextElement:
				forwardMessages.Elements = append(forwardMessages.Elements, splitPlainMessage(msgVal.Content)...)
			case *message.GroupImageElement:
				r.tryAppendImageByUrl(forwardMessages, msgVal.Url)
			case *message.AtElement:
				if msgVal.Target != 0 {
					forwardMessages.Append(message.NewText(fmt.Sprintf("@%v(%v)", msgVal.Display, msgVal.Target)))
				} else {
					forwardMessages.Append(message.NewText("@全体成员(转发)"))
				}
			case *message.FaceElement, *message.LightAppElement:
				forwardMessages.Append(msg)
			default:
				jsonBytes, _ := json.Marshal(msg)
				forwardMessages.Elements = append(forwardMessages.Elements, splitPlainMessage(fmt.Sprintf("%v\n", string(jsonBytes)))...)
			}

			// 如果加了该消息后会超出单个消息大小，则先放入待定队列
			jsonBytes, _ := json.Marshal(forwardMessages.Elements[len(forwardMessages.Elements)-1])
			msgSize += len(jsonBytes)
			// 需要确保每次至少转发一条消息
			if len(forwardMessages.Elements) > 1 && msgSize > maxMessageJsonSize {
				forwardMessages.Elements = forwardMessages.Elements[:len(forwardMessages.Elements)-1]
				msgSize -= len(jsonBytes)
				leftMessages.Append(msg)
			}
		}

		if !forRepeat {
			forwardMessages.Append(message.NewText(fmt.Sprintf(""+
				"\n"+
				"------------------------------\n"+
				"转发自 群[%v:%v] QQ[%v:%v] 时间[%v]",
				m.GroupName, m.GroupCode,
				m.Sender.Nickname, m.Sender.Uin,
				r.currentTime(),
			)))
		}

		forwardMessagesList = append(forwardMessagesList, forwardMessages)
		messages = leftMessages.Elements
	}

	return forwardMessagesList
}

func (r *QQRobot) onMemberJoin(m *client.MemberJoinGroupEvent, rule *Rule) error {
	config := rule.Config
	nowStr := r.currentTime()

	groupId := m.Group.Code
	newMemberUin := m.Member.Uin

	if _, ok := config.GroupIds[groupId]; !ok {
		return nil
	}

	// 判断该规则是否处理入群消息
	if !rule.Config.SendOnJoin {
		return nil
	}

	// ok

	// 回复消息
	replies := message.NewSendingMessage()

	// at该新入群成员
	replies.Append(message.NewAt(newMemberUin))

	// 发送指引信息
	replies.Append(message.NewText(config.GuideContent))

	// 如配置了图片url，则额外发送图片
	if config.ImageUrl != "" {
		r.tryAppendImageByUrl(replies, config.ImageUrl)
	}

	if len(replies.Elements) != 0 {
		rspId := r.cqBot.SendGroupMessage(groupId, replies)
		if rspId == -1 {
			logger.Error("【ReplyFail】", nowStr, p(m), rspId)
			return fmt.Errorf("reply fail, rspId=%v", rspId)
		}
		logger.Info("【OK】", nowStr, p(m.Group), 0, (replies), rspId)
	}

	if muteTime := config.MuteTime; muteTime != 0 {
		err := m.Member.Mute(uint32(muteTime))
		if err != nil {
			logger.Error("【Mute On Join Fail】", nowStr, config.Name, p(m), err)
		} else {
			logger.Info("【Mute On Join OK】", nowStr, config.Name, p(m), (replies))
		}
	}

	return nil
}

func (r *QQRobot) onPrivateOrTempMessage(senderFriendUin int64, tempGroupId int64, tempUin int64, m interface{}) {
	// 回复消息
	replies := message.NewSendingMessage()

	// 发送指引信息
	cfg := r.Config.Robot
	if cfg.PersonalMessageNotSupportedMessage != "" {
		replies.Append(message.NewText(cfg.PersonalMessageNotSupportedMessage))
	}

	if cfg.PersonalMessageNotSupportedImage != "" {
		r.tryAppendImageByUrl(replies, cfg.PersonalMessageNotSupportedImage)
	}

	if len(replies.Elements) == 0 {
		return
	}

	var rspId int32
	if senderFriendUin != 0 {
		rspId = r.cqBot.SendPrivateMessage(senderFriendUin, 0, replies)
	} else {
		rspId = r.cqBot.SendPrivateMessage(tempUin, tempGroupId, replies)
	}

	if rspId == -1 {
		logger.Error("【ReplyFail】", p(m), rspId)
		return
	}
	logger.Info("【OK】", p(m), 0, replies, rspId)
}
