set -x

# 拉取代码
git pull

# 构建新版本
go build -v -o qq_robot .

# 拉取私有仓库子模块的最新版本，并将最新配置复制到项目根目录
git submodule update --recursive --remote
cp qqrobot/setting/{config.toml,config.yml,device.json,session.token} .

# 运行
./qq_robot faststart
