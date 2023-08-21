//go:build !windows
// +build !windows

package applescript

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/crc-org/crc/v2/test/extended/util"
)

func ExecuteApplescript(scriptFilename string, args ...string) error {
	command := strings.Join(append(
		append([]string{"osascript"}, scriptFilename),
		args...),
		" ")
	return util.ExecuteCommandSucceedsOrFails(command, "succeeds")
}

func ExecuteApplescriptReturnShouldMatch(expectedOutput string,
	scriptFilename string, args ...string) error {
	command := strings.Join(append(
		append([]string{"osascript"}, scriptFilename),
		args...),
		" ")
	err := util.ExecuteCommand(command)
	if err != nil {
		return err
	}
	return util.CommandReturnShouldMatch("stdout", expectedOutput)
}

func GetScriptsPath(scriptsRelativePath string) (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if ok {
		return string(path.Dir(filename) +
			string(filepath.Separator) +
			scriptsRelativePath), nil
	}
	return "", fmt.Errorf("error recovering required resources for applescript installer handler")
}
