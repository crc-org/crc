// +build windows

package installer

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/RedHatQE/gowinx/pkg/win32/ux"
	"github.com/code-ready/crc/pkg/crc/logging"
)

const (
	installerWindowTitle string = "CodeReady Containers Setup"

	installerStartTime time.Duration = 20 * time.Second
	elementClickTime   time.Duration = 2 * time.Second
	installationTime   time.Duration = 75 * time.Second
)

var installFlow = []element{
	{"Next", elementClickTime, ux.BUTTON},
	{"I accept the terms in the License Agreement", elementClickTime, ux.CHECKBOX},
	{"Next", elementClickTime, ux.BUTTON},
	{"Next", elementClickTime, ux.BUTTON},
	{"Install", installationTime, ux.BUTTON},
	{"Finish", elementClickTime, ux.BUTTON}}

var rebootButton = element{"Yes", elementClickTime, ux.BUTTON}

type element struct {
	id          string
	delay       time.Duration
	elementType string
}

type gowinxHandler struct {
	CurrentUserPassword *string
	InstallerPath       *string
}

func NewInstaller(currentUserPassword, installerPath *string) Installer {
	// TODO check parameters as they are mandatory otherwise exit
	return gowinxHandler{
		CurrentUserPassword: currentUserPassword,
		InstallerPath:       installerPath}

}

func RequiredResourcesPath() (string, error) {
	return "", nil
}

func (g gowinxHandler) Install() error {
	// Initialize context
	initialize()
	if err := runInstaller(*g.InstallerPath); err != nil {
		return err
	}
	time.Sleep(installerStartTime)
	for _, action := range installFlow {
		// delay to get window as active
		// need to get window everytime as its layout is changing
		installer, err := ux.GetActiveElement(installerWindowTitle, ux.WINDOW)
		if err != nil {
			return err
		}
		element, err := installer.GetElement(action.id, action.elementType)
		if err != nil {
			err = fmt.Errorf("error getting %s with error %v", action.id, err)
			logging.Error(err)
			return err
		}
		if err := element.Click(); err != nil {
			err = fmt.Errorf("error clicking %s with error %v", action.id, err)
			logging.Error(err)
			return err
		}
		// Delay after action
		time.Sleep(action.delay)
	}
	// TODO which should be executed from a new cmd (to ensure ENVs)
	// Finalize context
	finalize()
	return nil
}

func (g gowinxHandler) RebootRequired() error {
	initialize()
	installer, err := ux.GetActiveElement(installerWindowTitle, ux.WINDOW)
	if err != nil {
		return err
	}
	element, err := installer.GetElement(rebootButton.id, rebootButton.elementType)
	if err != nil {
		return err
	}
	if err := element.Click(); err != nil {
		err = fmt.Errorf("error clicking %s with error %v", rebootButton.id, err)
		logging.Error(err)
		return err
	}
	finalize()
	return nil
}

func runInstaller(installerPath string) error {
	cmd := exec.Command("msiexec.exe", "/i", installerPath, "/qf")
	if err := cmd.Start(); err != nil {
		logging.Error(err)
	}
	// delay to get window as active
	time.Sleep(1 * time.Second)
	return nil
}

func initialize() {
	// Initialize context
	ux.Initialize()
	// Ensure clean desktop
	if err := exec.Command("powershell.exe", "-c", "(New-Object", "-ComObject", "\"Shell.Application\").minimizeall()").Run(); err != nil {
		logging.Error(err)
	}
}

func finalize() {
	// Finalize context
	ux.Finalize()
}
