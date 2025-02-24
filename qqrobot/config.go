package qqrobot

import (
	"os"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	logger "github.com/sirupsen/logrus"
)

// TencentAiChatAPI 腾讯ai聊天接口
const TencentAiChatAPI = "tbp.tencentcloudapi.com"

const (
	templateargsWorktime          = "$work_time$"           // 本次工作时长
	templateargsFoodname          = "$food_name$"           // 食物名字
	templateargsFoodpage          = "$food_page$"           // 食物页码
	templateargsCurrentperiodname = "$current_period_name$" // 当前时间段的名称
	templateargsMutetime          = "$mute_time$"           // 禁言时间
	templateargsCd                = "$cd$"                  // CD
	templateargsGitversion        = "$git_version$"         // 代码版本，若对应规则配置了changelog的链接，则会将这个变量替换为解析出的最新的版本号，如https://github.com/fzls/djc_helper/blob/master/CHANGELOG.MD
	templateargsUpdatemessage     = "$update_message$"      // 最新更新信息，若对应规则配置了changelog的链接，则会将这个变量替换为解析出的最新的更新信息，如https://github.com/fzls/djc_helper/blob/master/CHANGELOG.MD
	templateargsRealMoney         = "$real_money$"          // 结算金额
	templateargsSettleTime        = "$settle_time$"         // 对应结算操作的时间
)

// NotifyConfig 通知规则
type NotifyConfig struct {
	Name         string  `toml:"name"`          // 操作名称
	NotifyGroups []int64 `toml:"notify_groups"` // 通知的群
	Message      string  `toml:"message"`       // 通知的消息
}

// RobotConfig 机器人配置
type RobotConfig struct {
	IgnoreRobotQQs                     []int64      `toml:"ignore_robot_qqs"`                       // 机器人的qq号列表，这些号的消息将被忽略，避免多个机器人同时在群里不停循环回复
	Timeout                            int64        `toml:"timeout"`                                // http请求超时
	Debug                              bool         `toml:"debug"`                                  // 是否是调试模式
	OnStart                            NotifyConfig `toml:"on_start"`                               // 机器人上线时的操作
	OnStop                             NotifyConfig `toml:"on_stop"`                                // 机器人下线时的操作，参数：$work_time$=本次工作时长
	MaxRetryTimes                      int          `toml:"max_retry_times"`                        // 单条消息处理失败后，最多重试次数
	MaxContinueEmptyLines              int          `toml:"max_continue_empty_lines"`               // 最大允许的连续空行数目，为0则不限制
	TencentAiAppID                     string       `toml:"tencent_ai_app_id"`                      // 腾讯ai开放平台的应用ID，具体可见 https://console.cloud.tencent.com/tbp/bots
	TencentAiAppKey                    string       `toml:"tencent_ai_app_key"`                     // 腾讯ai开放平台的应用秘钥
	TencentAiBotID                     string       `toml:"tencent_ai_bot_id"`                      // 腾讯ai开放平台的机器人BotId
	ChatAnswerNotFoundMessage          string       `toml:"chat_answer_not_found_message"`          // 聊天结果未找到时的提示语
	PersonalMessageNotSupportedMessage string       `toml:"personal_message_not_supported_message"` // 不支持私聊时的提示语 # 本QQ是机器人，基本不会登录该QQ人工查看消息，如果有事，请私聊大号~
	PersonalMessageNotSupportedImage   string       `toml:"personal_message_not_supported_image"`   // 不支持私聊时的图片
	PprofListenAddr                    string       `toml:"pprof_listen_addr"`                      // pprof http监听地址
	EnableSellCard                     bool         `toml:"enable_sell_card"`                       // 是否启用卖卡功能
	SellCardEndTime                    string       `toml:"sell_card_end_time"`                     // 本次卖卡过期时间 %Y-%m-%d
}

// RuleType 规则类型
type RuleType string

// const (
// 	RuleType_AutoReply = "自动回复"
// 	RuleType_Command   = "机器人指令"
// 	RuleType_AtSomeOne = "AT某人"
// 	RuleType_Food      = "深夜美食"
// 	RuleType_Test      = "测试"
// )

// RuleTypeMaxApplyCountInfinite 不限制应用次数上限
const RuleTypeMaxApplyCountInfinite = -1

// RuleTypeConfig 规则类别
type RuleTypeConfig struct {
	Type          RuleType `toml:"type"`                 // 规则类别
	MaxApplyCount int32    `toml:"type_max_apply_count"` // 同一条消息最多应用该类型的规则的数目，-1表示不限制
}

// GroupTypeConfig 群类别
type GroupTypeConfig struct {
	Type     string  `toml:"type"`      // 群类别
	GroupIds []int64 `toml:"group_ids"` // 归属该类别的群组id列表
}

// ActionType 操作类型
type ActionType string

const (
	actionTypeGuide             ActionType = "guide"
	actionTypeCommand           ActionType = "command"
	actionTypeFood              ActionType = "food"
	actiontypeAichat            ActionType = "ai_chat"
	actiontypeSendupdatemessage ActionType = "send_update_message"
	actiontypeRepeater          ActionType = "repeater"
)

var (
	commandregexAddwhitelist = regexp.MustCompile(`\s*AddWhiteList\s+(?P<RuleName>.+?)\s+(?P<QQ>\d+)`)
	commandregexRulenamelist = regexp.MustCompile(`RuleNameList`)
	commandregexBuycard      = regexp.MustCompile(`\s*我想要给(?P<QQ>\d+)买一张(?P<CardIndex>[1-3]-[1-4])`)
	commandregexQuerycard    = regexp.MustCompile(`\s*给我康康现在还有哪些卡`)
	commandRegexMusic        = regexp.MustCompile(`.*(点歌|听歌|点首|来首|想听|要听)\s*(?P<MusicName>.+)`)
)

// RuleConfig 规则配置
type RuleConfig struct {
	Name                        string             `toml:"name"`                            // 规则名称
	Type                        RuleType           `toml:"type"`                            // 规则类别
	RawGroupIds                 []int64            `toml:"group_ids"`                       // 适用的QQ群ID列表
	GroupTypes                  []string           `toml:"group_types"`                     // 适用的QQ群类别，将于QQ群ID列表合并组成最终生效QQ群列表
	GroupIds                    map[int64]struct{} `toml:"-"`                               //
	RawKeywords                 []string           `toml:"keywords"`                        // 适用的关键词列表
	KeywordRegexes              []*regexp.Regexp   `toml:"-"`                               // 适用的关键词的正则表达式列表
	RawExcludeKeywords          []string           `toml:"exclude_keywords"`                // 需要过滤的关键词列表
	ExcludeKeywordRegexes       []*regexp.Regexp   `toml:"-"`                               // 需要过滤的关键词的正则表达式列表
	AtQQs                       []int64            `toml:"at_qqs"`                          // 需要判定at的qq列表
	ExcludeQQs                  []int64            `toml:"exclude_qqs"`                     // 排除的QQ列表
	ExcludeAdmin                bool               `toml:"exclude_admin"`                   // 是否排除管理员
	Action                      ActionType         `toml:"action"`                          // 动作
	SendOnJoin                  bool               `toml:"send_on_join"`                    // 是否在入群时发送
	AtQQsOnTrigger              []int64            `toml:"at_qqs_on_trigger"`               // 当触发该规则时，需要at的qq列表
	AtAllOnTrigger              bool               `toml:"at_all_on_trigger"`               // 当触发该规则时，是否需要@全体成员
	GuideContent                string             `toml:"guide_content"`                   // 内容
	ImageURL                    string             `toml:"image_url"`                       // 图片URL，若有，则会额外附加图片
	ImageURLList                []string           `toml:"image_url_list"`                  // 图片URL列表，若有，则会额外附加图片
	RandomImageUrls             []string           `toml:"random_image_urls"`               // 若配置，则从中随机一个作为图片发送，同时ImageUrl配置会被覆盖
	CD                          int64              `toml:"cd"`                              // cd时长（秒），0表示不设定，若设定，在cd内触发规则时，若设置了cd内回复内容，则回复该内容，否则视为未触发
	GuideContentInCD            string             `toml:"guide_content_in_cd"`             // cd内触发规则时的回复内容
	ForwardToQQs                []int64            `toml:"forward_to_qqs"`                  // 将消息转发到该QQ列表
	ForwardToGroups             []int64            `toml:"forward_to_groups"`               // 将消息转发到该QQ群列表
	RepeatToGroups              []int64            `toml:"repeat_to_groups"`                // 将消息复读到该QQ群列表
	RepeatToGroupTypes          []string           `toml:"repeat_to_group_types"`           // 复读适用的QQ群类别，将于QQ群ID列表合并组成最终生效QQ群列表
	FoodSiteURLList             []string           `toml:"food_site_url_list"`              // 美食图片来源网站列表
	FoodDescription             string             `toml:"food_description"`                // 美食描述，参数：$food_name$=食物名字
	RevokeMessage               bool               `toml:"revoke_message"`                  // 是否撤回该条消息
	MuteTime                    int64              `toml:"mute_time"`                       // 禁言时间，为0则表示不禁言(单位为秒)
	ParseMuteTime               bool               `toml:"parse_mute_time"`                 // 是否从消息从解析想要被禁言的时间
	TimePeriods                 []TimePeriod       `toml:"time_periods"`                    // 适用该规则的时间段（前者包含，后者不包含）
	TriggerRuleCount            int64              `toml:"trigger_rule_count"`              // TriggerRuleDuration内触发的规则数目是否超过该数目
	TriggerRuleDuration         int64              `toml:"trigger_rule_duration"`           // 判定恶意触发机器人规则的时间周期（秒）
	GitChangelogPage            string             `toml:"git_changelog_page"`              // 某git仓库的changelog的url，若设定，则将请求这个网页，从中解析出最新的版本号和更新信息，并替换到GuideContent中的$git_version$和$update_message$
	GuideContentHasPermission   string             `toml:"guide_content_has_permission"`    // 当有权限触发该指令时的回复
	GuideContentHasNoPermission string             `toml:"guide_content_has_no_permission"` // 当无权限触发该指令时的回复
}

// TimePeriod 时间规则
type TimePeriod struct {
	// 以下任意字段不设置则不检查
	StartSecond  int `toml:"start_second"`  // 起始的秒（包含），0-59
	EndSecond    int `toml:"end_second"`    // 截止的秒（包含），0-59
	StartMinute  int `toml:"start_minute"`  // 起始的分钟（包含），0-59
	EndMinute    int `toml:"end_minute"`    // 截止的分钟（包含），0-59
	StartHour    int `toml:"start_hour"`    // 起始的小时（包含），0-23
	EndHour      int `toml:"end_hour"`      // 截止的小时（包含），0-23
	StartWeekDay int `toml:"start_weekday"` // 起始的小时（包含），1-7表示周一到周日
	EndWeekDay   int `toml:"end_weekday"`   // 截止的小时（包含），1-7表示周一到周日
}

// MiscConfig 其他配置
type MiscConfig struct {
	Fireworks FireworksConfig `toml:"fireworks"`
	Ocr       OcrConfig       `toml:"ocr"`
}

// FireworksConfig 烟花活动
type FireworksConfig struct {
	Enable       bool    `toml:"enable"`
	Tips         string  `toml:"tips"`
	Image        string  `toml:"image"`
	NotifyGroups []int64 `toml:"notify_groups"`
}

// OcrConfig ocr配置
type OcrConfig struct {
	Enable bool `toml:"enable"`
}

// NotifyUpdateConfig 通知更新配置
type NotifyUpdateConfig struct {
	CheckInterval int64              `toml:"check_interval"` // 检查更新的间隔（秒）
	Rules         []NotifyUpdateRule `toml:"rules"`          // 检查规则
}

// NotifyUpdateRule 通知更新规则
type NotifyUpdateRule struct {
	Name             string   `toml:"name"`               // 名称
	NotifyGroups     []int64  `toml:"notify_groups"`      // 通知的群
	NotifyGroupTypes []string `toml:"notify_group_types"` // 通知适用的QQ群类别，将于QQ群ID列表合并组成最终生效QQ群列表
	Message          string   `toml:"message"`            // 通知的消息，参数：$git_version$=最新版本, $update_message$=更新信息
	ImageURL         string   `toml:"image_url"`          // 图片URL，若有，则会额外附加图片
	GitChangelogPage string   `toml:"git_changelog_page"` // git仓库的changelog的url，将请求这个网页，从中解析出最新的版本号和更新信息，并替换到message中的$git_version$和$update_message$
	AtQQsOnTrigger   []int64  `toml:"at_qqs_on_trigger"`  // 需要at的qq列表
	AtAllOnTrigger   bool     `toml:"at_all_on_trigger"`  // 是否需要@全体成员
}

// NotifySettleConfig 通知结算配置
type NotifySettleConfig struct {
	CheckInterval int64  `toml:"check_interval"` // 检查更新的间隔（秒），建议3600s
	APIUrl        string `toml:"api_url"`        // 结算api的url
	NotifyQQ      int64  `toml:"notify_qq"`      // 要通知的QQ
	StartMessage  string `toml:"start_message"`  // 开始结算的消息，参数：$real_money$=结算金额, $settle_time$=对应结算操作的时间
	FinishMessage string `toml:"end_message"`    // 完成结算的消息，参数：$real_money$=结算金额, $settle_time$=对应结算操作的时间
}

// Config 配置
type Config struct {
	Robot            RobotConfig        `toml:"robot"`
	Rules            []RuleConfig       `toml:"rules"`
	RuleTypeConfigs  []RuleTypeConfig   `toml:"rule_types"`
	GroupTypeConfigs []GroupTypeConfig  `toml:"rule_type_configs"`
	Misc             MiscConfig         `toml:"misc"`
	NotifyUpdate     NotifyUpdateConfig `toml:"notify_update"`
	NotifySettle     NotifySettleConfig `toml:"notify_settle"`
}

// LoadConfig 读取配置
func LoadConfig(configPath string) Config {
	// 读取配置
	var config Config
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		logger.Fatalf("load toml file fail, err=%v", err)
	}
	config.Init()
	logger.Debugf("%#v", config)

	return config
}

// Init 初始化配置中的一些字段
func (c *Config) Init() {
	for idx := range c.Rules {
		rule := &c.Rules[idx]

		for _, keyword := range rule.RawKeywords {
			rule.KeywordRegexes = append(rule.KeywordRegexes, regexp.MustCompile(keyword))
		}
		for _, keyword := range rule.RawExcludeKeywords {
			rule.ExcludeKeywordRegexes = append(rule.ExcludeKeywordRegexes, regexp.MustCompile(keyword))
		}

		rule.GroupIds = map[int64]struct{}{}
		for _, groupID := range rule.RawGroupIds {
			rule.GroupIds[groupID] = struct{}{}
		}
		for _, groupType := range rule.GroupTypes {
			for _, groupTypeCfg := range c.GroupTypeConfigs {
				if groupTypeCfg.Type != groupType {
					continue
				}

				for _, groupID := range groupTypeCfg.GroupIds {
					rule.GroupIds[groupID] = struct{}{}
				}
			}
		}

		rule.RepeatToGroups = c.mergeGroupTypesIntoGroups(rule.RepeatToGroups, rule.RepeatToGroupTypes)
	}

	for idx := range c.NotifyUpdate.Rules {
		rule := &c.NotifyUpdate.Rules[idx]

		rule.NotifyGroups = c.mergeGroupTypesIntoGroups(rule.NotifyGroups, rule.NotifyGroupTypes)
	}

	if err := c.check(); err != nil {
		logger.Errorf("Check failed, err=%v", err)
		os.Exit(-1)
	}
}

func (c *Config) mergeGroupTypesIntoGroups(groups []int64, groupTypes []string) (merged []int64) {
	merged = append(merged, groups...)

	for _, groupType := range groupTypes {
		for _, groupTypeCfg := range c.GroupTypeConfigs {
			if groupTypeCfg.Type != groupType {
				continue
			}

			for _, groupID := range groupTypeCfg.GroupIds {
				if InRangeInt64(groupID, merged) {
					continue
				}
				merged = append(merged, groupID)
			}
		}
	}

	return merged
}

// InRangeInt64 是否在列表内
func InRangeInt64(target int64, list []int64) bool {
	for _, value := range list {
		if value == target {
			return true
		}
	}

	return false
}

func (c *Config) check() error {
	for _, rule := range c.Rules {
		if rule.Type == "" {
			return errors.Errorf("rule=%v type=%v type not set", rule.Name, rule.Type)
		}
		exists := false
		for _, ruleType := range c.RuleTypeConfigs {
			if rule.Type == ruleType.Type {
				exists = true
				break
			}
		}
		if !exists {
			return errors.Errorf("rule=%v type=%v not valid", rule.Name, rule.Type)
		}
	}

	return nil
}
