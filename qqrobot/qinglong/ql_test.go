package qinglong

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getQlDir(t *testing.T) {
	assert.Equal(t, testQlDir, GetQlDir())

	_ = os.Unsetenv("QL_DIR")
	assert.Equal(t, defaultQlDir, GetQlDir())

	_ = os.Setenv("QL_DIR", testQlDir)
	assert.Equal(t, testQlDir, GetQlDir())
}

func Test_getPath(t *testing.T) {
	assert.Equal(t, testQlDir+"/test.txt", getPath("test.txt"))
}
