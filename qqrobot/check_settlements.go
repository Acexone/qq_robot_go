package qqrobot

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/Mrs4s/MiraiGo/message"
	"github.com/pkg/errors"
	logger "github.com/sirupsen/logrus"
)

// SettleResponse 结算结果
type SettleResponse struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Pid  int          `json:"pid"`
	Key  string       `json:"key"`
	Type interface{}  `json:"type"`
	Data []SettleInfo `json:"data"`
}

// SettleInfo 单条结算记录
type SettleInfo struct {
	ID             string      `json:"id"`
	UID            string      `json:"uid"`
	Batch          interface{} `json:"batch"`
	Auto           string      `json:"auto"`
	Type           string      `json:"type"`
	Account        string      `json:"account"`
	UserName       string      `json:"username"`
	Money          string      `json:"money"`
	RealMoney      string      `json:"realmoney"`
	AddTime        string      `json:"addtime"`
	EndTime        string      `json:"endtime"`
	Status         string      `json:"status"`
	TransferStatus string      `json:"transfer_status"`
	TransferResult interface{} `json:"transfer_result"`
	TransferDate   interface{} `json:"transfer_date"`
	Result         interface{} `json:"result"`
}

func (r *QQRobot) checkSettlements() {
	// 每天大概00:00:01的时候开始结算，12:00:00后某个时间完成结算
	now := time.Now()
	h, m, s := now.Clock()
	todaySettleStartTime := now.Add(-time.Duration(3600*h+60*m+s) * time.Second).Truncate(time.Second)
	todaySettleFinishTime := todaySettleStartTime.Add(12 * time.Hour)

	// 下列情况需要尝试处理
	// 1. 过了开始结算时间，但是今天尚未开始结算
	// 2. 过了完成结算时间，但是今天尚未完成结算
	needProcess := now.After(todaySettleStartTime) && r.lastSettleStartTime.Before(todaySettleStartTime) ||
		now.After(todaySettleFinishTime) && r.lastSettleFinishTime.Before(todaySettleStartTime)
	if !needProcess {
		return
	}

	settleInfo, err := r.getLatestSettleInfo()
	if err != nil {
		return
	}

	if r.lastSettleStartTime.Before(todaySettleStartTime) && settleInfo.AddTime != "" {
		startTime, _ := time.Parse("2006-01-02 15:04:05", settleInfo.AddTime)
		if startTime.Before(todaySettleStartTime) {
			// 不是今天的结算记录
			return
		}

		// 发送结算消息
		r.sendSettleMessage(r.Config.NotifySettle.StartMessage, settleInfo.RealMoney, settleInfo.AddTime)

		// 更新通知时间
		r.lastSettleStartTime = startTime
	}

	if r.lastSettleFinishTime.Before(todaySettleStartTime) && settleInfo.EndTime != "" {
		endTime, _ := time.Parse("2006-01-02 15:04:05", settleInfo.EndTime)
		if endTime.Before(todaySettleStartTime) {
			// 不是今天的结算记录
			return
		}

		// 发送结算消息
		r.sendSettleMessage(r.Config.NotifySettle.FinishMessage, settleInfo.RealMoney, settleInfo.EndTime)

		// 更新通知时间
		r.lastSettleFinishTime = endTime
	}
}

func (r *QQRobot) sendSettleMessage(templateMsg string, realMoney string, settleTime string) {
	reply := message.NewSendingMessage()
	msg := templateMsg
	msg = strings.ReplaceAll(msg, templateargsRealMoney, realMoney)
	msg = strings.ReplaceAll(msg, templateargsSettleTime, settleTime)
	reply.Append(message.NewText(msg))
	r.cqBot.SendPrivateMessage(r.Config.NotifySettle.NotifyQQ, 0, reply)
	logger.Infof("发送结算消息= %v", message.ToReadableString(reply.Elements))
}

// getLatestSettleInfo 获取最近的结算信息
func (r *QQRobot) getLatestSettleInfo() (*SettleInfo, error) {
	resp, err := r.httpClient.Get(r.Config.NotifySettle.APIUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytesData, _ := io.ReadAll(resp.Body)

	var settleResponse SettleResponse
	err = json.Unmarshal(bytesData, &settleResponse)
	if err != nil {
		return nil, err
	}

	if len(settleResponse.Data) == 0 {
		return nil, errors.Errorf("empty result")
	}

	return &settleResponse.Data[0], nil
}
