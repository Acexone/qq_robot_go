set -x

# 拉取代码
git pull

# 构建新版本
go build -v -o qq_robot .

# 从私有仓库子模块拉取最新配置，并复制到项目根目录
git submodule update --recursive
cp qqrobot/setting/{config.toml, config.yml, device.json} .

# 运行
./qq_robot faststart
