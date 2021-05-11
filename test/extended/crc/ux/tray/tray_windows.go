// +build windows

package tray

import (
	"fmt"

	clicumber "github.com/code-ready/clicumber/testsuite"
	crc "github.com/code-ready/crc/test/extended/crc/cmd"
)

const (
	trayAssemblyName string = "crc-tray.exe"
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
	if err := crc.SetConfigPropertyToValueSucceedsOrFails("enable-experimental-features", "true", "succeeds"); err != nil {
		return err
	}
	return clicumber.ExecuteCommandSucceedsOrFails("crc setup", "succeeds")
}

func (h handler) IsInstalled() error {
	command := fmt.Sprintf("tasklist /NH /FI \"IMAGENAME eq %s\"", trayAssemblyName)
	output := fmt.Sprintf("%s*", trayAssemblyName)
	return clicumber.CommandReturnShouldContain(command, output)
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

func (h handler) SetPullSecretFileLocation() error {
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
