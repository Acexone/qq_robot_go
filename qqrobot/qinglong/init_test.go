package qinglong

import "os"

var test_ql_dir = "test_data"

func init() {
	_ = os.Setenv("QL_DIR", test_ql_dir)
}
