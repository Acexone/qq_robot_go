package qqrobot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/Mrs4s/MiraiGo/message"
	"github.com/pkg/errors"
)

// 2021/10/02 5:25 by fzls

var errNoNewFood = errors.Errorf("tryFetchMoreFoodImages failed, cannot get new food image that has never been sent")

// FoodImageRegex 食物图片正则
var FoodImageRegex = regexp.MustCompile(`<img src="//(.*?\.jpg).*alt="(.*?)"`)

// MeiShiChinaResponse 美食中国回包
type MeiShiChinaResponse struct {
	Error int               `json:"error"`
	Data  []MeiShiChinaFood `json:"data"`
}

// MeiShiChinaFood 美食中国的食物
type MeiShiChinaFood struct {
	UID            string `json:"uid"`
	Username       string `json:"username"`
	ID             string `json:"id"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	Mainingredient string `json:"mainingredient"`
	Dateline       string `json:"dateline"`
	Subject        string `json:"subject"`
	Fcover         string `json:"fcover"`
	Cover          string `json:"cover"`
	Mpic           string `json:"mpic"`
	Tvpic          string `json:"tvpic"`
	Mscover        string `json:"mscover"`
	Path           string `json:"path"`
	Picname        string `json:"picname"`
	Collnum        string `json:"collnum"`
	Viewnum        int    `json:"viewnum"`
	Replynum       int    `json:"replynum"`
	Copyright      string `json:"copyright"`
	C320           string `json:"c320"`
	Avatar         string `json:"avatar"`
	Likenum        int    `json:"likenum"`
	Isfav          int    `json:"isfav"`
	Islike         int    `json:"islike"`
	Wapurl         string `json:"wapurl"`
}

func (r *QQRobot) tryFetchMoreFoodImages(rule *Rule, foodSiteURL string) error {
	// 请求网站链接
	siteURL := strings.ReplaceAll(foodSiteURL, templateargsFoodpage, strconv.FormatInt(rule.SiteToFoodPage[foodSiteURL], 10))
	resp, err := r.httpClient.Get(siteURL)
	if err != nil {
		return errors.Errorf("get food site err=%v, siteURL=%v\n", err, siteURL)
	}
	defer resp.Body.Close()

	// 获取网页内容
	bytesData, _ := ioutil.ReadAll(resp.Body)

	// 解析出所有美食图片
	var foodImages []FoodImage

	switch {
	case strings.Contains(siteURL, "www.xinshipu.com"):

		// 心食谱
		htmlText := string(bytesData)
		matches := FoodImageRegex.FindAllStringSubmatch(htmlText, -1)
		for _, match := range matches {
			foodImages = append(foodImages, FoodImage{
				Name: match[2],
				URL:  fmt.Sprintf("https://%v", match[1]),
			})
		}
	case strings.Contains(siteURL, "home.meishichina.com"):

		// 美食中国
		var response MeiShiChinaResponse
		err := json.Unmarshal(bytesData, &response)
		if err != nil {
			return err
		}

		for _, food := range response.Data {
			foodImages = append(foodImages, FoodImage{
				Name: food.Title,
				URL:  food.Fcover,
			})
		}
	default:
		return errors.Errorf("未支持的食谱网：%v", siteURL)
	}

	newFetched := make([]FoodImage, 0, len(foodImages))
	for _, foodImage := range foodImages {
		// 跳过已经发送过的食物
		if _, sent := rule.SentFoodImages[foodImage.URL]; sent {
			continue
		}

		// 缓存到内存中
		rule.CachedFoodImages[foodImage] = struct{}{}
		newFetched = append(newFetched, foodImage)
	}

	// 判断是否获取到了新的食物
	if len(newFetched) == 0 {
		return errNoNewFood
	}

	logger.Infof("tryFetchMoreFoodImages fetched %v new food, siteURL=%v, detail=%v", len(newFetched), siteURL, newFetched)

	return nil
}

func (r *QQRobot) createFoodMessage(rule *Rule) (messages *message.SendingMessage, err error) {
	// 看看还没有备货，没了就尝试去获取一次，失败了就放弃
	if len(rule.CachedFoodImages) == 0 {
		// 最多尝试3次
		for i := 0; i < 3; i++ {
			// 随机挑选一个食谱网站
			foodSiteURL := rule.Config.FoodSiteURLList[rand.Intn(len(rule.Config.FoodSiteURLList))]

			rule.UpdateFoodPage(foodSiteURL)
			err = r.tryFetchMoreFoodImages(rule, foodSiteURL)
			if err == nil {
				break
			}
		}
		if err != nil {
			return nil, err
		}
	}

	foodImage := r.getOneCachedFoodImage(rule)

	// 从缓存中移除并标记已发送
	delete(rule.CachedFoodImages, foodImage)
	rule.SentFoodImages[foodImage.URL] = struct{}{}

	// 发送食物到对应群聊中
	description := rule.Config.FoodDescription
	// 替换时间段
	description = strings.ReplaceAll(description, templateargsCurrentperiodname, getCurrentPeriodName())
	// 替换食物名参数
	description = strings.ReplaceAll(description, templateargsFoodname, foodImage.Name)

	messages = message.NewSendingMessage()
	messages.Append(message.NewText(description))
	r.tryAppendImageByURL(messages, foodImage.URL)

	return messages, nil
}

func (r *QQRobot) getOneCachedFoodImage(rule *Rule) FoodImage {
	// 获取缓存的第一个食物图片
	var foodImage FoodImage
	takeNthFood := rand.Intn(len(rule.CachedFoodImages))
	var idx int
	for fi := range rule.CachedFoodImages {
		if idx != takeNthFood {
			idx++
			continue
		}
		foodImage = fi
		break
	}
	logger.Infof("%v select the %v th food from %v foods, food=%v, 发完这个，库存食物图片还剩%v",
		r.currentTime(), takeNthFood+1, len(rule.CachedFoodImages), foodImage, len(rule.CachedFoodImages)-1)
	return foodImage
}
