package os

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/code-ready/crc/pkg/crc/logging"
)

const (
	LINUX   OS = "linux"
	DARWIN  OS = "darwin"
	WINDOWS OS = "windows"
)

type OS string

func (t OS) String() string {
	return string(t)
}

func CurrentOS() OS {
	switch runtime.GOOS {
	case "windows":
		return WINDOWS
	case "darwin":
		return DARWIN
	case "linux":
		return LINUX
	}
	panic("Unexpected OS type")
}

// ReplaceEnv changes the value of an environment variable
// It drops the existing value and appends the new value in-place
func ReplaceEnv(variables []string, varName string, value string) []string {
	var result []string
	for _, e := range variables {
		pair := strings.Split(e, "=")
		if pair[0] != varName {
			result = append(result, e)
		} else {
			result = append(result, fmt.Sprintf("%s=%s", varName, value))
		}
	}

	return result
}

func CopyFileContents(src string, dst string, permission os.FileMode) error {
	logging.Debugf("Copying '%s' to '%s'", src, dst)
	srcFile, err := os.Open(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("[%v] Cannot open src file '%s'", err, src)
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, permission)
	if err != nil {
		return fmt.Errorf("[%v] Cannot create dst file '%s'", err, dst)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("[%v] Cannot copy '%s' to '%s'", err, src, dst)
	}

	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("[%v] Cannot sync '%s' to '%s'", err, src, dst)
	}

	return nil
}

func WriteFileIfContentChanged(path string, new_content []byte, perm os.FileMode) (bool, error) {
	old_content, err := ioutil.ReadFile(filepath.Clean(path))
	if (err == nil) && (bytes.Equal(old_content, new_content)) {
		return false, nil
	}
	/* Intentionally ignore errors, just try to write the file if we can't read it */

	err = ioutil.WriteFile(path, new_content, perm)
	if err != nil {
		return false, err
	}
	return true, nil
}
