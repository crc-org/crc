// +build windows

package installer

import (
	"fmt"

	goautoit "github.com/adrianriobo/goautoit"
)

type autoitHandler struct {
	CurrentUserPassword *string
	InstallerPath       *string
}

func NewInstaller(currentUserPassword *string, installerPath *string) Installer {
	// TODO check parameters as they are mandatory otherwise exit
	return autoitHandler{
		CurrentUserPassword: currentUserPassword,
		InstallerPath:       installerPath}

}

func RequiredResourcesPath() (string, error) {
	return "", nil
}

func (a autoitHandler) Install() error {
	command := fmt.Sprintf("msiexec.exe /i %s /qf", *a.InstallerPath)
	installerPid := goautoit.RunWait(command)
	if installerPid == 0 {
		return fmt.Errorf("error starting the msi installer")
	}
	goautoit.MouseClick("Next")
	return nil
}
