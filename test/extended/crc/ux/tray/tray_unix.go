// +build !windows

package tray

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	clicumber "github.com/code-ready/clicumber/testsuite"
	"github.com/code-ready/crc/test/extended/os/applescript"
	"github.com/code-ready/crc/test/extended/util"
)

const (
	scriptsRelativePath    string = "applescripts"
	checkTrayIconIsVisible string = "checkTrayIconIsVisible.applescript"
	clickTrayMenuItem      string = "clickTrayMenuItem.applescript"
	setPullSecret          string = "setPullSecret.applescript"
	getTrayFieldlValue     string = "getTrayFieldlValue.applescript"
	installTray            string = "installTray.applescript"
	getOCLoginCommand      string = "getOCLoginCommand.applescript"
	runOCLoginCommand      string = "runOCLoginCommand.applescript"

	bundleIdentifier string = "com.redhat.codeready.containers"
	appPath          string = "/Applications/CodeReady Containers.app"

	uxCheckAccessibilityDuration = "2s"
	uxCheckAccessibilityRetry    = 10
)

var (
	elements = map[string]string{
		actionStart:   "start",
		actionStop:    "stop",
		actionDelete:  "delete",
		actionQuit:    "quit",
		fieldState:    "cluster_status",
		userKubeadmin: "kubeadmin_login",
		userDeveloper: "developer_login"}
)

type applescriptHandler struct {
	bundleLocation         *string
	pullSecretFileLocation *string
}

func NewTray(bundleLocationValue, pullSecretFileLocationValue *string) Tray {
	if runtime.GOOS == "darwin" {
		return applescriptHandler{
			bundleLocation:         bundleLocationValue,
			pullSecretFileLocation: pullSecretFileLocationValue}

	}
	return nil
}

func RequiredResourcesPath() (string, error) {
	return applescript.GetScriptsPath(scriptsRelativePath)
}

func (a applescriptHandler) Install() error {
	err := clicumber.ExecuteCommandSucceedsOrFails("crc setup", "succeeds")
	if err != nil {
		return err
	}
	// Required to pass parameters with spaces to applescript
	sanitizedAppPath := strings.Join(append([]string{"\""}, appPath, "\""), "")
	return applescript.ExecuteApplescript(installTray, sanitizedAppPath)
}

func (a applescriptHandler) IsInstalled() error {
	return executeCommandSucceeds("launchctl list | grep crc", "0.*tray")
}

func (a applescriptHandler) IsAccessible() error {
	return checkAccessible(func() error {
		return applescript.ExecuteApplescript(
			checkTrayIconIsVisible, bundleIdentifier)
	}, "Tray icon")
}

func (a applescriptHandler) ClickStart() error {
	return clickButtonByAction(actionStart)
}

func (a applescriptHandler) ClickStop() error {
	return clickButtonByAction(actionStop)
}

func (a applescriptHandler) ClickDelete() error {
	return clickButtonByAction(actionDelete)
}

func (a applescriptHandler) ClickQuit() error {
	return clickButtonByAction(actionQuit)
}

func (a applescriptHandler) SetPullSecret() error {
	return applescript.ExecuteApplescript(
		setPullSecret, bundleIdentifier, *a.pullSecretFileLocation)
}

func (a applescriptHandler) IsClusterRunning() error {
	return util.MatchWithRetry(stateRunning, checkTrayShowsStatusValue,
		trayClusterStateRetries, trayClusterStateTimeout)
}

func (a applescriptHandler) IsClusterStopped() error {
	return util.MatchWithRetry(stateStopped, checkTrayShowsStatusValue,
		trayClusterStateRetries, trayClusterStateTimeout)
}

func (a applescriptHandler) CopyOCLoginCommandAsKubeadmin() error {
	return clickCopyOCLoginCommand(userKubeadmin)
}

func (a applescriptHandler) CopyOCLoginCommandAsDeveloper() error {
	return clickCopyOCLoginCommand(userDeveloper)
}

func (a applescriptHandler) ConnectClusterAsKubeadmin() error {
	return applescript.ExecuteApplescriptReturnShouldMatch(
		userKubeadmin, runOCLoginCommand)
}

func (a applescriptHandler) ConnectClusterAsDeveloper() error {
	return applescript.ExecuteApplescriptReturnShouldMatch(
		userDeveloper, runOCLoginCommand)
}

func clickButtonByAction(actionName string) error {
	return clickOnElement(actionName, clickTrayMenuItem)
}

func clickCopyOCLoginCommand(userName string) error {
	return clickOnElement(userName, getOCLoginCommand)
}

func clickOnElement(elementName string, scriptName string) error {
	element, err := getElement(elementName, elements)
	if err != nil {
		return err
	}
	return applescript.ExecuteApplescript(
		scriptName, bundleIdentifier, element)
}

func checkTrayShowsStatusValue(expectedValue string) error {
	element, err := getElement(fieldState, elements)
	if err != nil {
		return err
	}
	return applescript.ExecuteApplescriptReturnShouldMatch(
		expectedValue, getTrayFieldlValue, bundleIdentifier, element)
}

func checkAccessible(uxIsAccessible func() error, component string) error {
	retryDuration, err := time.ParseDuration(uxCheckAccessibilityDuration)
	if err != nil {
		return err
	}
	for i := 0; i < uxCheckAccessibilityRetry; i++ {
		err := uxIsAccessible()
		if err == nil {
			return nil
		}
		time.Sleep(retryDuration)
	}
	return fmt.Errorf("%s is not accessible", component)
}

// TODO review which helper use
func executeCommandSucceeds(command string, expectedOutput string) error {
	err := clicumber.ExecuteCommand(command)
	if err != nil {
		return err
	}
	return clicumber.CommandReturnShouldMatch("stdout", expectedOutput)
}
