// +build windows

package installer

import (
	"fmt"
)

type handler struct {
	currentUserPassword *string
	installerPath       *string
}

func NewInstaller(currentUserPassword *string, installerPath *string) Installer {
	// TODO check parameters as they are mandatory otherwise exit
	return handler{
		currentUserPassword: currentUserPassword,
		installerPath:       installerPath}

}

func RequiredResourcesPath() (string, error) {
	return "", nil
}

func (h handler) Install() error {
	return fmt.Errorf("not implemented yet")
}
