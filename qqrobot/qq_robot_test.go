package qqrobot

import (
	"fmt"
	"reflect"
	"testing"

	mylogger "github.com/fzls/logger"
)

// 2020/10/30 17:39 by fzls

func Test_convertChineseNumber(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		// 正确的例子
		{name: "", args: "一", want: "1"},
		{name: "", args: "十", want: "10"},
		{name: "", args: "十一", want: "11"},
		{name: "", args: "二十", want: "20"},
		{name: "", args: "二十一", want: "21"},
		{name: "", args: "二百", want: "200"},
		{name: "", args: "三百零四", want: "304"},
		{name: "", args: "五百六十七", want: "567"},
		{name: "", args: "八千九百一十二", want: "8912"},
		{name: "", args: "八千一十二", want: "8012"},
		{name: "", args: "八千零三", want: "8003"},
		{name: "", args: "八千零二十五", want: "8025"},
		{name: "", args: "给我来个八千零二十五秒禁言套餐", want: "给我来个8025秒禁言套餐"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertChineseNumber(tt.args); got != tt.want {
				t.Errorf("convertChineseNumber(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func Test_getLatestGitVersion(t *testing.T) {
	// 初始化日志
	var err error
	logger, err = mylogger.NewLogger("logs", "qq_robot", mylogger.DebugLevel)
	if err != nil {
		fmt.Printf("new logger err=%v\n", err)
		return
	}

	qqRobot := NewQQRobot(nil, "../config.toml")
	version, updateMessage := qqRobot.getLatestGitVersion("https://github.com/fzls/djc_helper/blob/master/CHANGELOG.MD")
	t.Logf("version=%v, updateMessage如下：\n%v", version, updateMessage)
}

func Test_version_to_version_int_list(t *testing.T) {
	tests := []struct {
		version string
		want    []int64
	}{
		{version: "1.2.3", want: []int64{1, 2, 3}},
		{version: "v1.2.3", want: []int64{1, 2, 3}},
		{version: "v1.12.3", want: []int64{1, 12, 3}},
		{version: "v1.0.3", want: []int64{1, 0, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := versionToVersionIntList(tt.version); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("version_to_version_int_list(%v) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func Test_version_less(t *testing.T) {
	tests := []struct {
		version_left  string
		version_right string
		want          bool
	}{
		{version_left: "v1.0.0", version_right: "v1.0.0", want: false},
		{version_left: "v1.0.0", version_right: "v1.0.1", want: true},
		{version_left: "v1.0.0", version_right: "v1.1.0", want: true},
		{version_left: "v1.0.0", version_right: "v2.0.0", want: true},
		{version_left: "v2.0.0", version_right: "v1.0.0", want: false},
		{version_left: "v2.0.0", version_right: "v10.0.0", want: true},
		{version_left: "v3.2.9.1", version_right: "v4.0.0", want: true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v-%v", tt.version_left, tt.version_right), func(t *testing.T) {
			if got := versionLess(tt.version_left, tt.version_right); got != tt.want {
				t.Errorf("version_less(%v, %v) = %v, want %v", tt.version_left, tt.version_right, got, tt.want)
			}
		})
	}
}
