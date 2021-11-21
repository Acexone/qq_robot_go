set -x

git pull

go build -v -o qq_robot .

./qq_robot faststart
