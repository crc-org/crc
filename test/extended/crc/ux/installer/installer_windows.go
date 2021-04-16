// +build windows

package installer

import (
	"fmt"
)

type handler struct {
	CurrentUserPassword *string
	InstallerPath       *string
}

func NewInstaller(currentUserPassword *string, installerPath *string) Installer {
	// TODO check parameters as they are mandatory otherwise exit
	return handler{
		BundleLocation: bundleLocationValue,
		PullSecretFile: pullSecretFileValue}

}

func RequiredResourcesPath() (string, error) {
	return "", nil
}

func (h handler) Install() error {
	return fmt.Errorf("not implemented yet")
}
