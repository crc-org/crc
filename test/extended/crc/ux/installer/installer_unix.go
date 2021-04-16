// +build !windows

package installer

import (
	"runtime"

	"github.com/code-ready/crc/test/extended/os/applescript"
)

const (
	scriptsRelativePath string = "applescripts"
	installScript       string = "install.applescript"
)

type applescriptHandler struct {
	currentUserPassword *string
	installerPath       *string
}

func NewInstaller(currentUserPassword *string, installerPath *string) Installer {
	// TODO check parameters as they are mandatory otherwise exit
	if runtime.GOOS == "darwin" {
		return applescriptHandler{
			currentUserPassword: currentUserPassword,
			installerPath:       installerPath}

	}
	return nil
}

func RequiredResourcesPath() (string, error) {
	return applescript.GetScriptsPath(scriptsRelativePath)
}

func (a applescriptHandler) Install() error {
	return applescript.ExecuteApplescript(
		installScript,
		*a.installerPath,
		*a.currentUserPassword)
}
