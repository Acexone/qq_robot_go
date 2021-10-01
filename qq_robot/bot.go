package qq_robot

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/coolq"
)

type QQRobot struct {
	cqBot *coolq.CQBot
}

func NewQQRobot(cqRobot *coolq.CQBot) *QQRobot {
	r := &QQRobot{
		cqBot: cqRobot,
	}

	return r
}

func (r *QQRobot) RegisterHandlers() {
	//r.cqBot.Client.OnPrivateMessage(rprivateMessageEvent)
	r.cqBot.Client.OnGroupMessage(r.OnGroupMessage)
	//r.cqBot.Client.OnSelfPrivateMessage(rprivateMessageEvent)
	//r.cqBot.Client.OnSelfGroupMessage(rgroupMessageEvent)
	//r.cqBot.Client.OnTempMessage(rtempMessageEvent)
	//r.cqBot.Client.OnGroupMuted(rgroupMutedEvent)
	//r.cqBot.Client.OnGroupMessageRecalled(rgroupRecallEvent)
	//r.cqBot.Client.OnGroupNotify(rgroupNotifyEvent)
	//r.cqBot.Client.OnFriendNotify(rfriendNotifyEvent)
	//r.cqBot.Client.OnMemberSpecialTitleUpdated(rmemberTitleUpdatedEvent)
	//r.cqBot.Client.OnFriendMessageRecalled(rfriendRecallEvent)
	//r.cqBot.Client.OnReceivedOfflineFile(rofflineFileEvent)
	//r.cqBot.Client.OnJoinGroup(rjoinGroupEvent)
	//r.cqBot.Client.OnLeaveGroup(rleaveGroupEvent)
	//r.cqBot.Client.OnGroupMemberJoined(rmemberJoinEvent)
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

func (r *QQRobot) Demo() {
}

func (r *QQRobot) OnGroupMessage(client *client.QQClient, m *message.GroupMessage) {
}
