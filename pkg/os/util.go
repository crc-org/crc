package os

import (
	"fmt"
	"runtime"
	"strings"

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
