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
	Finished            chan bool
}

func NewInstaller(currentUserPassword, installerPath *string) Installer {
	// TODO check parameters as they are mandatory otherwise exit
	return gowinxHandler{
		CurrentUserPassword: currentUserPassword,
		InstallerPath:       installerPath,
		Finished:            make(chan bool)}

}

func RequiredResourcesPath() (string, error) {
	return "", nil
}

func (g gowinxHandler) Install() error {
	// Initialize context
	ux.Initialize()
	if err := runInstaller(*g.InstallerPath, g.Finished); err != nil {
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
	ux.Finalize()
	return nil
}

func (g gowinxHandler) RebootRequired() error {
	ux.Initialize()
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
	<-g.Finished
	ux.Finalize()
	return nil
}

func runInstaller(installerPath string, finished chan bool) error {
	// cmd := exec.Command("powershell.exe", "-c", "msiexec.exe", "/i", installerPath, "/qf")
	// if err := cmd.Start(); err != nil {
	// 	return fmt.Errorf("error starting %v with error %v", cmd, err)
	// }
	go openInstaller(installerPath, finished)
	// delay to get window as active
	time.Sleep(1 * time.Second)
	return nil
}

func openInstaller(installerPath string, finished chan bool) {
	// msiexecArguments := strings.Join(append([]string{"'"}, "/i", installerPath, "/qf", "'"), " ")
	// cmd := exec.Command("powershell.exe", "-c", "Start-Process", "msiexec.exe", "-Wait", "-ArgumentList", msiexecArguments)
	// // cmd := exec.Command("Start-Process", "msiexec.exe", "-Wait", "-ArgumentList", "/i", installerPath, "/qf")
	// // if err := cmd.Start(); err != nil {
	// // 	return fmt.Errorf("error starting %v with error %v", cmd, err)
	// // }
	// // cmd := exec.Command("powershell.exe", "-c", "msiexec.exe", "/i", installerPath, "/qf")
	cmd := exec.Command("msiexec.exe", "/i", installerPath, "/qf")
	if err := cmd.Start(); err != nil {
		logging.Error(err)
		finished <- true
	}
	if err := cmd.Wait(); err != nil {
		logging.Error(err)
	}
	finished <- true
}
