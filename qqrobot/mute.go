package qqrobot

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/Mrs4s/MiraiGo/message"
)

// 2021/10/02 5:23 by fzls

// MaxMuteTime 最大禁言30天
const MaxMuteTime = 30 * 60 * 60 * 24

func truncatMuteTime(muteTime int64) int64 {
	if muteTime > MaxMuteTime {
		return MaxMuteTime
	}

	return muteTime
}

var muteRegexMap = map[int64]*regexp.Regexp{
	1:                     regexp.MustCompile(`(\d+(\.\d*)?).*秒`),
	60:                    regexp.MustCompile(`(\d+(\.\d*)?).*分钟`),
	60 * 15:               regexp.MustCompile(`(\d+(\.\d*)?).*刻钟`),
	60 * 60:               regexp.MustCompile(`(\d+(\.\d*)?).*小时`),
	60 * 60 * 24:          regexp.MustCompile(`(\d+(\.\d*)?).*天`),
	60 * 60 * 24 * 7:      regexp.MustCompile(`(\d+(\.\d*)?).*周`),
	60 * 60 * 24 * 30:     regexp.MustCompile(`(\d+(\.\d*)?).*月`),
	60 * 60 * 24 * 30 * 4: regexp.MustCompile(`(\d+(\.\d*)?).*季度`),
	60 * 60 * 24 * 365:    regexp.MustCompile(`(\d+(\.\d*)?).*年`),
}

func parseMuteTime(messageChain []message.IMessageElement) int64 {
	for _, msg := range messageChain {
		if msgVal, ok := msg.(*message.TextElement); ok {
			// 处理中文数字
			text := convertChineseNumber(msgVal.Content)
			for baseMuteTime, muteRegex := range muteRegexMap {
				if match := muteRegex.FindStringSubmatch(text); match != nil {
					return int64(float64(baseMuteTime) * parseFloat64(match[1]))
				}
			}
		}
	}

	return 0
}

var traditionalChineseToSimplifiedChinese = map[string]string{
	"壹": "一", "贰": "二", "叁": "三", "肆": "四", "伍": "五", "陆": "六", "柒": "七", "捌": "八", "玖": "九", "拾": "十", "佰": "百", "仟": "千",
}

var chineseDigitToArabicDigitMap = map[string]string{
	"零": "0",
	"一": "1", "二": "2", "三": "3", "四": "4", "五": "5", "六": "6", "七": "7", "八": "8", "九": "9",
}

var chineseNumberUnits = []string{"十", "百", "千", "万"}
var arbicNumberUnits = []string{"0", "00", "000", "0000"}

// 并不正确的简单中文转阿拉伯数字，如多个单位（十百千）混用的情况会给出错误的结果
func convertChineseNumber(originText string) string {
	converted := originText
	// 首先繁体转简体
	for traditional, simplified := range traditionalChineseToSimplifiedChinese {
		if strings.Contains(converted, traditional) {
			converted = strings.ReplaceAll(converted, traditional, simplified)
		}
	}
	// 然后替换数字
	for chineseDigit, arabicDigit := range chineseDigitToArabicDigitMap {
		if strings.Contains(converted, chineseDigit) {
			converted = strings.ReplaceAll(converted, chineseDigit, arabicDigit)
		}
	}
	// 最后将必要的单位转为0，todo：这个算法是错误的，只能处理一些简单的情况。处理中间一个零代表多个0的情况，如五千零三
	for idx, unit := range chineseNumberUnits {
		if !strings.Contains(converted, unit) {
			continue
		}

		var replacedWith string

		utf8converted := []rune(converted)
		runeUnit := []rune(unit)[0]

		// 判断是否是首个单位
		unitIndex := 0
		for runeIdx, runeChar := range utf8converted {
			if runeChar == runeUnit {
				unitIndex = runeIdx
				continue
			}
		}

		// 判断这个单位右侧是否是一个阿拉伯数字
		if unitIndex+1 < len(utf8converted) && isDigit(byte(utf8converted[unitIndex+1])) {
			// 根据右侧数字数目，替换0
			rightDigitCount := 0
			for runeIdx := unitIndex + 1; runeIdx < len(utf8converted); runeIdx++ {
				if isDigit(byte(utf8converted[runeIdx])) {
					rightDigitCount++
				}
			}
			zeroCount := len(arbicNumberUnits[idx]) - rightDigitCount
			replacedWith = strings.Repeat("0", zeroCount)
		} else {
			// 否则，转为对应数目的0
			replacedWith = arbicNumberUnits[idx]
		}

		// 如果左边没有其他东西，则左边加个1，如十一 => 11
		if unitIndex == 0 {
			replacedWith = "1" + replacedWith
		}

		converted = strings.ReplaceAll(converted, unit, replacedWith)
	}
	return converted
}

func isDigit(char byte) bool {
	return '0' <= char && char <= '9'
}

func parseFloat64(str string) float64 {
	f, _ := strconv.ParseFloat(str, 64)
	return f
}
