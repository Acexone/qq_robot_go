set -x

# 拉取代码
git pull

# 构建新版本
go build -v -o qq_robot .

# 运行
./qq_robot faststart
