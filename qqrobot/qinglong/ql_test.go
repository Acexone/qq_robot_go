package qinglong

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getQlDir(t *testing.T) {
	assert.Equal(t, test_ql_dir, GetQlDir())

	_ = os.Unsetenv("QL_DIR")
	assert.Equal(t, default_ql_dir, GetQlDir())

	_ = os.Setenv("QL_DIR", test_ql_dir)
	assert.Equal(t, test_ql_dir, GetQlDir())
}

func Test_getPath(t *testing.T) {
	assert.Equal(t, test_ql_dir+"/test.txt", getPath("test.txt"))
}
