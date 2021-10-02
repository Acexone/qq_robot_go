# qq机器人

基于 [go-cqhttp](https://github.com/Mrs4s/go-cqhttp) 实现的qq机器人，在其上额外增加了一个直接本地处理消息并根据配置自动回复的逻辑, 具体改动可见 *qq_robot* 目录

# 使用说明

## 配置机器人

请参考 [go-cqhttp的文档](https://docs.go-cqhttp.org/)

## 配置自动回复逻辑

复制qq_robot/default_config.toml到qq_robot.exe所在目录，并重命名为config.toml 然后打开qq_robot/config.go和config.toml，按照注释，自行调整配置，并添加各种规则

# TODO

[ ] 看看是否有更多事件值得接入，目前仅处理了群聊、私聊和加群这三个事件