// Package qinglong 封装青龙相关逻辑
package qinglong

import (
	"os"
)

var default_ql_dir = "/qinglong/data"

func getQlDir() string {
	qlEnv := os.Getenv("QL_DIR")
	if qlEnv != "" {
		return qlEnv
	}
	return default_ql_dir
}

func getPath(path string) string {
	return getQlDir() + "/" + path
}
