set -x

# 拉取代码，包括子模块
git pull --recurse-submodules

# 从私有仓库子模块复制最新配置到项目根目录
cp qqrobot/setting/{config.toml,config.yml,device.json} .

# 构建新版本
go build -v -o qq_robot .

# 运行
./qq_robot faststart
