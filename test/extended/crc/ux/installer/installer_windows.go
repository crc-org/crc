// +build windows

package installer

import (
	"fmt"
	"time"

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
	installerPid := goautoit.Run(command)
	fmt.Printf("After run\n")
	if installerPid == 0 {
		return fmt.Errorf("error starting the msi installer")
	}
	fmt.Printf("Waiting for window\n")
	// Calling Sleep method
	// time.Sleep(8 * time.Second)

	goautoit.WinWait("CodeReady Containers Setup")
	fmt.Printf("After Waiting for window\n")
	windowActive := goautoit.WinActive("CodeReady Containers Setup")
	fmt.Printf("After run\n")
	if windowActive == 0 {
		return fmt.Errorf("error windows active 0")
	}
	fmt.Printf("Window is active\n")
	// goautoit.ControlClick("CodeReady Containers Setup", "&Next", "Button1")
	goautoit.ControlClick("CodeReady Containers Setup", "&Next", "Button1")
	time.Sleep(2 * time.Second)
	goautoit.ControlClick("CodeReady Containers Setup", "&accept", "Button1")
	time.Sleep(2 * time.Second)
	// goautoit.WinActive("CodeReady Containers Setup")
	goautoit.ControlClick("CodeReady Containers Setup", "&Next", "Button3")
	goautoit.ControlClick("CodeReady Containers Setup", "&Next", "Button1")
	time.Sleep(2 * time.Second)
	goautoit.ControlClick("", "&Install", "Button1")
	time.Sleep(30 * time.Second)
	goautoit.ControlClick("CodeReady Containers Setup", "&Finish", "Button1")
	//I &accept the terms in the License Agreement
	// fmt.Printf("After click\n")
	// TODO which should be executed from a new cmd (to ensure ENVs)
	return nil
}
