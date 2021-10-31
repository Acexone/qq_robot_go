package qqrobot

import (
	"math/rand"
	"strings"

	logger "github.com/sirupsen/logrus"
)

// 2020/06/01 2:42 by fzls

// MaxFoodPageMap 美食网站与其最大页数
var MaxFoodPageMap = map[string]int64{
	"www.xinshipu.com":     30,
	"home.meishichina.com": 40,
}

// FoodImage 美食图片
type FoodImage struct {
	Name string
	URL  string
}

// Rule 规则，附带一些缓存数据
type Rule struct {
	Config            RuleConfig
	ProcessedMessages map[int64]struct{}
	SiteToFoodPage    map[string]int64              // 最新食谱页面的页码数
	SiteToFetchedPage map[string]map[int64]struct{} // 当前已经抓取过的食谱页码
	CachedFoodImages  map[FoodImage]struct{}        // 目前缓存的食物图片信息集合
	SentFoodImages    map[string]struct{}           // 目前已发送过的食物的图片url的集合
}

// UpdateFoodPage 更新食物页数
func (r *Rule) UpdateFoodPage(foodSiteURL string) {
	if r.SiteToFetchedPage[foodSiteURL] == nil {
		r.SiteToFetchedPage[foodSiteURL] = map[int64]struct{}{}
	}
	fetchedPage := r.SiteToFetchedPage[foodSiteURL]

	var MaxFoodPage int64 = 30
	for site, maxFoodPage := range MaxFoodPageMap {
		if strings.Contains(foodSiteURL, site) {
			MaxFoodPage = maxFoodPage
		}
	}

	if len(fetchedPage) == int(MaxFoodPage) {
		return
	}

	for {
		foodPage := 1 + rand.Int63n(MaxFoodPage)
		if _, fetched := fetchedPage[foodPage]; !fetched {
			fetchedPage[foodPage] = struct{}{}
			r.SiteToFoodPage[foodSiteURL] = foodPage
			logger.Infof("rule=%v UpdateFoodPage to %v", r.Config.Name, foodPage)
			return
		}
	}
}

// NewRule 创建新的规则
func NewRule(config RuleConfig) *Rule {
	return &Rule{
		Config:            config,
		ProcessedMessages: map[int64]struct{}{},
		SiteToFoodPage:    map[string]int64{},
		SiteToFetchedPage: map[string]map[int64]struct{}{},
		CachedFoodImages:  map[FoodImage]struct{}{},
		SentFoodImages:    map[string]struct{}{},
	}
}
