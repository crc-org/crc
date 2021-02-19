package os

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"os/user"

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

func TestFileExists(t *testing.T) {
	dir, err := ioutil.TempDir("", "fileexists")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	filename := filepath.Join(dir, "testfile1")
	_, err = WriteFileIfContentChanged(filename, []byte("content"), 0644)
	assert.NoError(t, err)
	assert.True(t, FileExists(filename))

	_, err = WriteFileIfContentChanged(filename, []byte("newcontent"), 0000)
	assert.NoError(t, err)
	assert.True(t, FileExists(filename))

	dirname := filepath.Join(dir, "testdir")
	err = os.MkdirAll(dirname, 0700)
	assert.NoError(t, err)
	filename = filepath.Join(dirname, "testfile2")
	_, err = WriteFileIfContentChanged(filename, []byte("content"), 0644)
	assert.NoError(t, err)
	assert.True(t, FileExists(filename))

	err = os.Chmod(dirname, 0000)
	assert.NoError(t, err)
	if runtime.GOOS == "windows" {
		assert.True(t, FileExists(filename))
	} else {
		user, _ := user.Current()
		if user != nil && user.Uid != "0" {
			assert.False(t, FileExists(filename))
		} else {
			/* If the user is root, chmod 000 $dir won't block file existence checks */
			assert.True(t, FileExists(filename))
		}
	}

	filename = filepath.Join(dirname, "nonexistent")
	assert.False(t, FileExists(filename))
}
