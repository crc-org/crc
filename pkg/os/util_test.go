package os

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceEnv(t *testing.T) {
	env := []string{"HOME=/home/user/", "PATH=/bin:/sbin:/usr/bin", "LC_ALL=de_DE.UTF8"}
	replaced := ReplaceOrAddEnv(env, "LC_ALL", "C")

	assert.Len(t, replaced, 3)
	assert.Equal(t, env[2], "LC_ALL=de_DE.UTF8")
	assert.Equal(t, replaced[2], "LC_ALL=C")
}

func TestAddEnv(t *testing.T) {
	env := []string{"HOME=/home/user/", "PATH=/bin:/sbin:/usr/bin", "LC_ALL=de_DE.UTF8"}
	replaced := ReplaceOrAddEnv(env, "KUBECONFIG", "some-data")

	assert.Len(t, replaced, 4)
	assert.Equal(t, replaced[3], "KUBECONFIG=some-data")
}

func TestFileContentFuncs(t *testing.T) {
	dir, err := ioutil.TempDir("", "filecontent")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	filename := filepath.Join(dir, "testfile")

	assert.Error(t, FileContentMatches(filename, []byte("aaaa")))

	written, err := WriteFileIfContentChanged(filename, []byte("aaaa"), 0600)
	assert.NoError(t, err)
	assert.Equal(t, written, true)

	written, err = WriteFileIfContentChanged(filename, []byte("aaaa"), 0600)
	assert.NoError(t, err)
	assert.Equal(t, written, false)

	assert.NoError(t, FileContentMatches(filename, []byte("aaaa")))
	assert.Error(t, FileContentMatches(filename, []byte("aaaaa")))

	written, err = WriteFileIfContentChanged(filename, []byte("aaaaa"), 0600)
	assert.NoError(t, err)
	assert.Equal(t, written, true)

	assert.Error(t, FileContentMatches(filename, []byte("aaaa")))
	assert.NoError(t, FileContentMatches(filename, []byte("aaaaa")))
}
