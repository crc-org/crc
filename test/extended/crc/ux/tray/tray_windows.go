// +build windows

package tray

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/RedHatQE/gowinx/pkg/win32/desktop/notificationarea"
	"github.com/RedHatQE/gowinx/pkg/win32/interaction"
	"github.com/RedHatQE/gowinx/pkg/win32/ux"
	clicumber "github.com/code-ready/clicumber/testsuite"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/test/extended/util"
)

type gowinxHandler struct {
	bundleLocation         *string
	pullSecretFileLocation *string
}

const (
	trayAssemblyName string = "crc-tray.exe"

	notificationIcon string = "CodeReady Containers"
	contextMenu      string = "menu"
	loginMenu        string = "loginMenu"

	pullSecretWindowID   string = "Pull Secret Picker"
	pullSecretButtonOkID string = "OK"

	elementClickTime time.Duration = 2 * time.Second
)

var (
	elements = map[string]string{
		actionStart:   "start",
		actionStop:    "stop",
		actionDelete:  "delete",
		actionQuit:    "exit",
		fieldState:    "status",
		loginMenu:     "copy-oc-login",
		userKubeadmin: "kubeadmin",
		userDeveloper: "developer"}
)

func NewTray(bundleLocationValue, pullSecretFileLocationValue *string) Tray {
	return gowinxHandler{
		bundleLocation:         bundleLocationValue,
		pullSecretFileLocation: pullSecretFileLocationValue}
}

func RequiredResourcesPath() (string, error) {
	return "", nil
}

func (g gowinxHandler) Install() error {
	return clicumber.ExecuteCommandSucceedsOrFails("crc setup", "succeeds")
}

func (g gowinxHandler) IsInstalled() (err error) {
	trayAssamblyNameFilter := fmt.Sprintf("IMAGENAME eq %s", trayAssemblyName)
	out, err := exec.Command("tasklist", "/NH", "/FI", trayAssamblyNameFilter).Output()
	if !strings.Contains(string(out), trayAssemblyName) {
		err = fmt.Errorf("%s process not found", trayAssemblyName)
	}
	return
}

func (g gowinxHandler) IsAccessible() (err error) {
	initialize()
	err = notificationarea.ShowHiddenNotificationArea()
	if err != nil {
		_, _, err = notificationarea.GetIconPositionByTitle(notificationIcon)
	}
	finalize()
	return
}

func (g gowinxHandler) ClickStart() error {
	return click(actionStart)
}

func (g gowinxHandler) ClickStop() error {
	return click(actionStop)
}

func (g gowinxHandler) ClickDelete() error {
	return click(actionDelete)
}

func (g gowinxHandler) ClickQuit() error {
	return click(actionQuit)
}

func (g gowinxHandler) SetPullSecret() (err error) {
	// Initialize base elements
	initialize()
	err = fillFullSecretField(*g.pullSecretFileLocation)
	// Finalize context
	finalize()
	return
}

func (g gowinxHandler) IsClusterRunning() error {
	return util.MatchWithRetry(stateRunning, checkTrayShowsStatusValue,
		trayClusterStateRetries, trayClusterStateTimeout)
}

func (g gowinxHandler) IsClusterStopped() error {
	return util.MatchWithRetry(stateStopped, checkTrayShowsStatusValue,
		trayClusterStateRetries, trayClusterStateTimeout)
}

func (g gowinxHandler) CopyOCLoginCommandAsKubeadmin() error {
	return clickOnSubmenu(loginMenu, userKubeadmin)
}

func (g gowinxHandler) CopyOCLoginCommandAsDeveloper() error {
	return clickOnSubmenu(loginMenu, userDeveloper)
}

func (g gowinxHandler) ConnectClusterAsKubeadmin() error {
	return connectClusterAs(userKubeadmin)
}

func (g gowinxHandler) ConnectClusterAsDeveloper() error {
	return connectClusterAs(userDeveloper)
}

func initialize() {
	// Initialize context
	ux.Initialize()
}

func finalize() {
	// Finalize context
	ux.Finalize()
}

func click(action string) (err error) {
	// Initialize base elements
	initialize()
	_, err = clickOnContextMenu(action)
	// Finalize context
	finalize()
	return
}

func clickOnSubmenu(submenuElement, action string) (err error) {
	// Initialize base elements
	initialize()
	if submenu, clickErr := clickOnContextMenu(submenuElement); clickErr != nil {
		err = clickErr
	} else {
		_, err = clickAction(action, submenu)
	}
	// Finalize context
	finalize()
	return
}

func clickOnContextMenu(element string) (*ux.UXElement, error) {
	contextMenu, err := getContextMenu()
	if err != nil {
		return nil, err
	}
	return clickAction(element, contextMenu)
}

func clickAction(element string, menu *ux.UXElement) (*ux.UXElement, error) {
	buttonID, err := getElement(element, elements)
	if err != nil {
		return nil, err
	}
	button, err := menu.GetElement(buttonID, ux.MENUITEM)
	if err != nil {
		return nil, err
	}
	if err := button.Click(); err != nil {
		return button, err
	}
	time.Sleep(elementClickTime)
	return button, nil
}

func getContextMenu() (*ux.UXElement, error) {
	err := notificationarea.ShowHiddenNotificationArea()
	if err != nil {
		return nil, err
	}
	x, y, err := notificationarea.GetIconPositionByTitle(notificationIcon)
	if err != nil {
		return nil, err
	}
	if err := interaction.Click(int32(x), int32(y)); err != nil {
		return nil, err
	}
	time.Sleep(elementClickTime)
	return ux.GetActiveElement(contextMenu, ux.MENU)
}

func connectClusterAs(connectedUser string) error {
	//  Get oc
	err := clicumber.ExecuteCommand("crc oc-env | Invoke-Expression")
	if err != nil {
		return err
	}
	// Copy command from clipboard
	err = clicumber.ExecuteCommand("$clipboardCommand=Get-Clipboard")
	if err != nil {
		return err
	}
	// Run copied command
	err = clicumber.ExecuteCommand("Invoke-Expression -Command $clipboardCommand")
	if err != nil {
		return err
	}
	// Clear
	err = clicumber.ExecuteCommand("clear")
	if err != nil {
		return err
	}
	// Check user
	return clicumber.CommandReturnShouldContain("oc whoami", connectedUser)
}

func checkTrayShowsStatusValue(expectedValue string) (err error) {
	// Initialize base elements
	initialize()
	contextMenu, errContextMenu := getContextMenu()
	if errContextMenu != nil {
		err = errContextMenu
	} else {
		if statusLabel, errStatusLabel := contextMenu.GetElementByType(ux.TEXT); errStatusLabel != nil {
			err = errStatusLabel
		} else if statusLabel.GetName() != expectedValue {
			err = fmt.Errorf("Tray does not show value %s", expectedValue)
		}
	}
	// Finalize context
	finalize()
	return
}

func fillFullSecretField(pullSecretLocation string) error {
	time.Sleep(elementClickTime)
	// Get pull secret content
	err := exec.Command("powershell.exe", "-c", "Get-Content", pullSecretLocation, "|", "Set-Clipboard").Run()
	if err != nil {
		return err
	}
	pullSecretWindow, err := ux.GetActiveElement(pullSecretWindowID, ux.WINDOW)
	if err != nil {
		logging.Infof("Pullsecret file already configured")
		return nil
	}
	pullSecretField, err := pullSecretWindow.GetElementByType(ux.EDIT)
	if err != nil {
		return err
	}

	pullSecretContent, err := exec.Command("powershell.exe", "-c", "Get-Clipboard").Output()
	if err != nil {
		return err
	}

	err = pullSecretField.SetValue(string(pullSecretContent))
	if err != nil {
		return err
	}
	okButton, err := pullSecretWindow.GetElement(pullSecretButtonOkID, ux.BUTTON)
	if err != nil {
		return err
	}
	if err := okButton.Click(); err != nil {
		return err
	}
	time.Sleep(elementClickTime)
	return nil
}
