// Package qinglong 封装青龙相关逻辑
package qinglong

import (
	"os"
)

var defaultQlDir = "/root/qinglong/data"

// GetQlDir 获取青龙数据目录路径
func GetQlDir() string {
	qlEnv := os.Getenv("QL_DIR")
	if qlEnv != "" {
		return qlEnv
	}
	return defaultQlDir
}

func getPath(path string) string {
	return GetQlDir() + "/" + path
}
