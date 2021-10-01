# qq机器人
基于 [go-cqhttp](https://github.com/Mrs4s/go-cqhttp) 实现的qq机器人，在其上额外增加了一个直接本地处理消息并根据配置自动回复的逻辑, 具体改动可见 *qq_robot* 目录

# 使用说明
## 配置机器人
请参考 [go-cqhttp的文档](https://docs.go-cqhttp.org/)

## 配置自动回复逻辑
请打开config.toml，按照注释和现有例子，自行调整配置