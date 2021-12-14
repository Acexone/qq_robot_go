package qinglong

import "os"

var testQlDir = "test_data"

func init() {
	_ = os.Setenv("QL_DIR", testQlDir)
}
