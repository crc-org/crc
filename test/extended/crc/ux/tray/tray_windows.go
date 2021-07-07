// +build windows

package tray

import (
	"fmt"
)

type handler struct {
	bundleLocation         *string
	pullSecretFileLocation *string
}

func NewTray(bundleLocationValue *string, pullSecretFileLocationValue *string) Tray {
	return handler{
		bundleLocation:         bundleLocationValue,
		pullSecretFileLocation: pullSecretFileLocationValue}

}

func RequiredResourcesPath() (string, error) {
	return "", nil
}

func (h handler) Install() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) IsInstalled() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) IsAccessible() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) ClickStart() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) ClickStop() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) ClickDelete() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) ClickQuit() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) SetPullSecret() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) IsInitialStatus() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) IsClusterRunning() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) IsClusterStopped() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) CopyOCLoginCommandAsKubeadmin() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) CopyOCLoginCommandAsDeveloper() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) ConnectClusterAsKubeadmin() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) ConnectClusterAsDeveloper() error {
	return fmt.Errorf("not implemented yet")
}
