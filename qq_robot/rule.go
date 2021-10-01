package qq_robot

import (
	"math/rand"
	"strings"
)

// 2020/06/01 2:42 by fzls

var MAX_FOOD_PAGE_MAP = map[string]int64{
	"www.xinshipu.com":     30,
	"home.meishichina.com": 40,
}

type FoodImage struct {
	Name string
	Url  string
}

type Rule struct {
	Config            RuleConfig
	ProcessedMessages map[int64]struct{}
	SiteToFoodPage    map[string]int64              // 最新食谱页面的页码数
	SiteToFetchedPage map[string]map[int64]struct{} // 当前已经抓取过的食谱页码
	CachedFoodImages  map[FoodImage]struct{}        // 目前缓存的食物图片信息集合
	SentFoodImages    map[string]struct{}           // 目前已发送过的食物的图片url的集合
}

func (r *Rule) UpdateFoodPage(foodSiteUrl string) {
	if r.SiteToFetchedPage[foodSiteUrl] == nil {
		r.SiteToFetchedPage[foodSiteUrl] = map[int64]struct{}{}
	}
	fetchedPage := r.SiteToFetchedPage[foodSiteUrl]

	var MAX_FOOD_PAGE int64 = 30
	for site, maxFoodPage := range MAX_FOOD_PAGE_MAP {
		if strings.Contains(foodSiteUrl, site) {
			MAX_FOOD_PAGE = maxFoodPage
		}
	}

	if len(fetchedPage) == int(MAX_FOOD_PAGE) {
		return
	}

	for {
		foodPage := 1 + rand.Int63n(MAX_FOOD_PAGE)
		if _, fetched := fetchedPage[foodPage]; !fetched {
			fetchedPage[foodPage] = struct{}{}
			r.SiteToFoodPage[foodSiteUrl] = foodPage
			logger.Debugf("rule=%v UpdateFoodPage to %v", r.Config.Name, foodPage)
			return
		}
	}
}

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
