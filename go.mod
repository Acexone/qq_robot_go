module github.com/Mrs4s/go-cqhttp

go 1.18

replace github.com/Mrs4s/go-cqhttp => github.com/fzls/qq_robot_go v1.0.0-beta8

// 魔改后需要额外引入的依赖项，单独列出，避免后面又冲突
require (
	github.com/BurntSushi/toml v1.1.0
	github.com/fzls/logger v1.1.1
	github.com/gookit/color v1.5.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.290
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tbp v1.0.290
)

require github.com/xo/terminfo v0.0.0-20210125001918-ca9a967f8778 // indirect

// 以下为go-cqhttp原本的依赖
require (
	github.com/Microsoft/go-winio v0.5.1
	github.com/Mrs4s/MiraiGo v0.0.0-20220405134734-9cb9e80d99d8
	github.com/RomiChan/syncx v0.0.0-20220320130821-c88644afda9c
	github.com/RomiChan/websocket v1.4.3-0.20220123145318-307a86b127bc
	github.com/fumiama/go-hide-param v0.1.4
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/mattn/go-colorable v0.1.12
	github.com/pkg/errors v0.9.1
	github.com/segmentio/asm v1.1.3
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	github.com/tidwall/gjson v1.14.0
	github.com/wdvxdr1123/go-silk v0.0.0-20210316130616-d47b553def60
	go.mongodb.org/mongo-driver v1.8.3
	golang.org/x/crypto v0.0.0-20211215153901-e495a2d5b3d3
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

require (
	github.com/RomiChan/protobuf v0.0.0-20220318113238-d8a99598f896 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.5.3 // indirect
	github.com/fumiama/imgsz v0.0.2 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/lestrrat-go/strftime v1.0.5 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/onsi/gomega v1.18.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.11 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.0.2 // indirect
	github.com/xdg-go/stringprep v1.0.2 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/net v0.0.0-20220412020605-290c469a71a5 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20220422013727-9388b58f7150 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/xerrors v0.0.0-20220411194840-2f41105eb62f // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	modernc.org/libc v1.8.1 // indirect
	modernc.org/mathutil v1.2.2 // indirect
	modernc.org/memory v1.0.4 // indirect
)
