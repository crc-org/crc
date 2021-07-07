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

	trayClusterStateTimeout = "90"
)

var (
	elements = []Element{
		{
			Name:         actionStart,
			AXIdentifier: "start"},
		{
			Name:         actionStop,
			AXIdentifier: "stop"},
		{
			Name:         actionDelete,
			AXIdentifier: "delete"},
		{
			Name:         actionQuit,
			AXIdentifier: "quit"},
		{
			Name:         fieldState,
			AXIdentifier: "cluster_status"},
		{
			Name:         userKubeadmin,
			AXIdentifier: "kubeadmin_login"},
		{
			Name:         userDeveloper,
			AXIdentifier: "developer_login"},
	}
)

type applescriptHandler struct {
	bundleLocation         *string
	pullSecretFileLocation *string
}

func NewTray(bundleLocationValue *string, pullSecretFileLocationValue *string) Tray {
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

func (a applescriptHandler) IsInitialStatus() error {
	return waitTrayShowsFieldWithValue(fieldState, stateInitial)
}

func (a applescriptHandler) IsClusterRunning() error {
	return waitTrayShowsFieldWithValue(fieldState, stateRunning)
}

func (a applescriptHandler) IsClusterStopped() error {
	return waitTrayShowsFieldWithValue(fieldState, stateStopped)
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
		scriptName, bundleIdentifier, element.AXIdentifier)
}

func waitTrayShowsFieldWithValue(field string, expectedValue string) error {
	retryCount := 15
	iterationDuration, extraDuration, err :=
		util.GetRetryParametersFromTimeoutInSeconds(retryCount, trayClusterStateTimeout)
	if err != nil {
		return err
	}
	for i := 0; i < retryCount; i++ {
		err := checkTrayShowsFieldWithValue(field, expectedValue)
		if err == nil {
			return nil
		}
		time.Sleep(iterationDuration)
	}
	if extraDuration != 0 {
		time.Sleep(extraDuration)
	}
	return fmt.Errorf("Tray did not showed %s ", expectedValue)
}

func checkTrayShowsFieldWithValue(field string, expectedValue string) error {
	element, err := getElement(field, elements)
	if err != nil {
		return err
	}
	return applescript.ExecuteApplescriptReturnShouldMatch(
		expectedValue, getTrayFieldlValue, bundleIdentifier, element.AXIdentifier)
}

func getElement(name string, elements []Element) (Element, error) {
	for _, e := range elements {
		if name == e.Name {
			return e, nil
		}
	}
	return Element{},
		fmt.Errorf("element '%s', Can not be accessed from the tray", name)
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
