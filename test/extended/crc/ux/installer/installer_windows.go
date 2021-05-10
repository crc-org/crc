// +build windows

package installer

import (
	"fmt"
	"time"

	goautoit "github.com/adrianriobo/goautoit"
)

const (
	installerWindowTitle string        = "CodeReady Containers Setup"
	elementClickDelay    time.Duration = 2 * time.Second
	installationTime     time.Duration = 30 * time.Second
)

type installerElement struct {
	name   string
	id     string
	screen string
}

var (
	welcomeNextButton         = installerElement{name: "Next", id: "1", screen: "welcome"}
	licenseAcceptCheck        = installerElement{name: "accept", id: "1", screen: "license"}
	licenseNextButton         = installerElement{name: "Next", id: "3", screen: "license"}
	destinantionNextButton    = installerElement{name: "Next", id: "1", screen: "destination"}
	installationInstallButton = installerElement{name: "Install", id: "1", screen: "installation"}
	finalizationFinishButton  = installerElement{name: "Finish", id: "1", screen: "finalization"}
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
	if err := runInstaller(*a.InstallerPath); err != nil {
		return err
	}
	// Welcome screen
	if err := clickButton(welcomeNextButton); err != nil {
		return err
	}
	// License screen
	if err := clickButton(licenseAcceptCheck); err != nil {
		return err
	}
	if err := clickButton(licenseNextButton); err != nil {
		return err
	}
	// Destination
	if err := clickButton(destinantionNextButton); err != nil {
		return err
	}
	// Installation
	if err := clickButton(installationInstallButton); err != nil {
		return err
	}
	// wait installation process
	time.Sleep(installationTime)
	// Finalization
	if err := clickButton(finalizationFinishButton); err != nil {
		return err
	}
	// TODO which should be executed from a new cmd (to ensure ENVs)
	return nil
}

func runInstaller(installerPath string) error {
	command := fmt.Sprintf("msiexec.exe /i %s /qf", installerPath)
	installerPid := goautoit.Run(command)
	if installerPid == 0 {
		return fmt.Errorf("error starting the msi installer")
	}
	goautoit.WinWait(installerWindowTitle)
	return nil
}

func clickButton(element installerElement) error {
	// Ensure the installer is the active window
	windowActive := goautoit.WinActive(installerWindowTitle)
	if windowActive == 0 {
		return fmt.Errorf("error activating %s", installerWindowTitle)
	}
	if exitCode := goautoit.ControlClick("", "&"+element.name, "Button"+element.id); exitCode == 0 {
		return fmt.Errorf("error clicking on  %s button on screen %s", element.name, element.screen)
	}
	time.Sleep(elementClickDelay)
	return nil
}
