package qq_robot

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Mrs4s/MiraiGo/message"
)

// 2021/10/02 5:25 by fzls
func (r *QQRobot) processCommand(commandStr string, m *message.GroupMessage) (err error, msg string, extraReplies []message.IMessageElement) {
	var match []string
	if match = CommandRegex_AddWhiteList.FindStringSubmatch(commandStr); len(match) == len(CommandRegex_AddWhiteList.SubexpNames()) {
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
	} else if match = CommandRegex_RuleNameList.FindStringSubmatch(commandStr); len(match) == len(CommandRegex_RuleNameList.SubexpNames()) {
		for _, rule := range r.Rules {
			if _, ok := rule.Config.GroupIds[m.GroupCode]; !ok {
				continue
			}

			if len(msg) == 0 {
				msg += "规则集合："
			}
			msg += ", " + rule.Config.Name
		}
	} else if match = CommandRegex_BuyCard.FindStringSubmatch(commandStr); len(match) == len(CommandRegex_BuyCard.SubexpNames()) {
		now := time.Now()
		endTime, _ := time.Parse("2006-01-02", r.Config.Robot.SellCardEndTime)
		if !r.Config.Robot.EnableSellCard || now.After(endTime) {
			return nil, "目前尚未启用卖卡功能哦", nil
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
			return err, "", nil
		}

		json.Unmarshal(out, &msg)

		if strings.Contains(msg, "成功发送以下卡片") {
			image, err := r._makeLocalImage("https://z3.ax1x.com/2020/12/16/r1yWZT.png")
			if err == nil {
				extraReplies = append(extraReplies, image)
			}
		}
	} else if match = CommandRegex_QueryCard.FindStringSubmatch(commandStr); len(match) == len(CommandRegex_QueryCard.SubexpNames()) {
		logger.Infof("开始查询卡片信息~")
		cmd := exec.Command("python", "sell_cards.py",
			"--run_remote",
			"--query",
		)
		cmd.Dir = "D:\\_codes\\Python\\djc_helper_public"
		out, err := cmd.Output()

		if err != nil {
			return err, "", nil
		}

		json.Unmarshal(out, &msg)
	} else if match = CommandRegex_Music.FindStringSubmatch(commandStr); len(match) == len(CommandRegex_Music.SubexpNames()) {
		// full_match|musicName
		musicName := match[1]

		musicElem, err := r.makeMusicShareElement(musicName, message.QQMusic)
		if err != nil {
			return fmt.Errorf("没有找到歌曲：%v", musicName), "", nil
		}

		msg = fmt.Sprintf("请欣赏歌曲：%v", musicName)
		extraReplies = append(extraReplies, musicElem)
	} else {
		return fmt.Errorf("没有找到该指令哦"), "", nil
	}

	return nil, msg, extraReplies
}
