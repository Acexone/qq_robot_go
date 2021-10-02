package qq_robot

import (
	"strconv"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tbp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tbp/v20190627"
)

func (r *QQRobot) initAiChat() {
	cfg := r.Config.Robot

	if cfg.TencentAiAppId == "" || cfg.TencentAiAppKey == "" || cfg.TencentAiBotId == "" {
		logger.Warnf("未配置腾讯ai的appid、appkey、botid，将不初始化aichat，详情可见 https://console.cloud.tencent.com/tbp/bots")
		return
	}

	credential := common.NewCredential(cfg.TencentAiAppId, cfg.TencentAiAppKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = TencentAiChatApi
	r.aiClient, _ = tbp.NewClient(credential, "", cpf)
}

// 使用腾讯ai开放平台的 智能对话平台 TBP https://console.cloud.tencent.com/tbp/bots
func (r *QQRobot) aiChat(targetQQ int64, chatText string) (responseText string) {
	if r.aiClient == nil {
		return ""
	}

	cfg := r.Config.Robot

	request := tbp.NewTextProcessRequest()

	request.BotId = common.StringPtr(cfg.TencentAiBotId)
	request.BotEnv = common.StringPtr("release")
	request.TerminalId = common.StringPtr(strconv.FormatInt(targetQQ, 10))
	request.InputText = common.StringPtr(chatText)

	response, err := r.aiClient.TextProcess(request)
	if err != nil {
		logger.Debugf("aiChat(qq=%v, text=%v) err=%v", targetQQ, chatText, err)
		return
	}

	answer := strings.Builder{}
	for _, resGroup := range response.Response.ResponseMessage.GroupList {
		if resGroup.Content == nil {
			continue
		}
		answer.WriteString(*resGroup.Content)
	}

	return answer.String()
}
