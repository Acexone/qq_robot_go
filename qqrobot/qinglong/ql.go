// Package qinglong 封装青龙相关逻辑
package qinglong

import (
	"os"
)

func getQlDir() string {
	qlEnv := os.Getenv("QL_DIR")
	if qlEnv != "" {
		return qlEnv
	}
	return "/qinglong/data"
}

func getPath(path string) string {
	return getQlDir() + "/" + path
}
