package qqrobot

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/message"
	logger "github.com/sirupsen/logrus"
)

// 2021/10/02 5:25 by fzls
func (r *QQRobot) processCommand(commandStr string, m *message.GroupMessage) (msg string, extraReplies []message.IMessageElement, err error) {
	var match []string
	if match = commandregexAddwhitelist.FindStringSubmatch(commandStr); len(match) == len(commandregexAddwhitelist.SubexpNames()) {
		// full_match|ruleName|qq
		ruleName := match[1]
		qq, _ := strconv.ParseInt(match[2], 10, 64)
		for _, rule := range r.Rules {
			if ruleName != rule.Config.Name {
				continue
			}
			rule.Config.ExcludeQQs = append(rule.Config.ExcludeQQs, qq)
			logger.Info("【Command】", commandStr)

			if len(msg) != 0 {
				msg += " | "
			}
			msg += fmt.Sprintf("已将【%v】加入到规则【%v】的白名单", qq, ruleName)
		}
	} else if match = commandregexRulenamelist.FindStringSubmatch(commandStr); len(match) == len(commandregexRulenamelist.SubexpNames()) {
		for _, rule := range r.Rules {
			if _, ok := rule.Config.GroupIds[m.GroupCode]; !ok {
				continue
			}

			if len(msg) == 0 {
				msg += "规则集合："
			}
			msg += ", " + rule.Config.Name
		}
	} else if match = commandregexBuycard.FindStringSubmatch(commandStr); len(match) == len(commandregexBuycard.SubexpNames()) {
		now := time.Now()
		endTime, _ := time.Parse("2006-01-02", r.Config.Robot.SellCardEndTime)
		if !r.Config.Robot.EnableSellCard || now.After(endTime) {
			return "目前尚未启用卖卡功能哦", nil, nil
		}

		qq := match[1]
		cardIndex := match[2]

		logger.Infof("开始调用卖卡脚本~")
		cmd := exec.Command("python", "sell_cards.py",
			"--run_remote",
			"--target_qq", qq,
			"--card_index", cardIndex,
		)
		cmd.Dir = "D:\\_codes\\Python\\djc_helper_public"
		out, err := cmd.Output()

		if err != nil {
			return "", nil, err
		}

		err = json.Unmarshal(out, &msg)
		if err != nil {
			return "", nil, err
		}

		if strings.Contains(msg, "成功发送以下卡片") {
			image, err := r._makeLocalImage("https://z3.ax1x.com/2020/12/16/r1yWZT.png")
			if err == nil {
				extraReplies = append(extraReplies, image)
			}
		}
	} else if match = commandregexQuerycard.FindStringSubmatch(commandStr); len(match) == len(commandregexQuerycard.SubexpNames()) {
		logger.Infof("开始查询卡片信息~")
		cmd := exec.Command("python", "sell_cards.py",
			"--run_remote",
			"--query",
		)
		cmd.Dir = "D:\\_codes\\Python\\djc_helper_public"
		out, err := cmd.Output()

		if err != nil {
			return "", nil, err
		}

		err = json.Unmarshal(out, &msg)
		if err != nil {
			return "", nil, err
		}
	} else if match = commandRegexMusic.FindStringSubmatch(commandStr); len(match) == len(commandRegexMusic.SubexpNames()) {
		// full_match|听歌关键词|musicName
		musicName := match[2]

		musicElem, err := r.makeMusicShareElement(musicName, message.QQMusic)
		if err != nil {
			return "", nil, errors.Errorf("没有找到歌曲：%v", musicName)
		}

		msg = fmt.Sprintf("请欣赏歌曲：%v", musicName)
		extraReplies = append(extraReplies, musicElem)
	} else {
		return "", nil, errors.Errorf("没有找到该指令哦")
	}

	return msg, extraReplies, nil
}
