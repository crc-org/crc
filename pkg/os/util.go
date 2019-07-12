package os

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/kardianos/osext"
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

func CurrentExecutable() (string, error) {
	currentExec, err := osext.Executable()
	if err != nil {
		return "", err
	}
	return currentExec, nil
}

func CopyFileContents(src string, dst string, permission os.FileMode) error {
	logging.DebugF("Copying '%s' to '%s'\n", src, dst)
	srcFile, err := os.Open(src)
	defer srcFile.Close()
	if err != nil {
		return fmt.Errorf("[%v] Cannot open src file '%s'", err, src)
	}

	destFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, permission)
	defer destFile.Close()
	if err != nil {
		return fmt.Errorf("[%v] Cannot create dst file '%s'", err, dst)
	}

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

// GetFilePath returns the file path even it is symlink
func GetFilePath(path string) (string, error) {
	fInfo, err := os.Lstat(path)
	if err != nil {
		return path, err
	}
	if fInfo.Mode()&os.ModeSymlink != 0 {
		return os.Readlink(path)
	}
	return path, nil
}
