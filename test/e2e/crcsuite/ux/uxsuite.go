package ux

import (
	"flag"
	"fmt"
	"os"

	"github.com/code-ready/crc/test/extended/crc/cmd"
	"github.com/code-ready/crc/test/extended/crc/ux/installer"
	"github.com/code-ready/crc/test/extended/crc/ux/notification"
	"github.com/code-ready/crc/test/extended/crc/ux/tray"
	"github.com/code-ready/crc/test/extended/util"
	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"
)

var trayHandler tray.Tray
var installerHandler installer.Installer
var notificationHandler notification.Notification
var currentUserPassword string
var installerPath string

const (
	clusterStateWaitTimeout int = 200
	clusterStateWaitRetries int = 10
)

// FeatureContext defines godog.Suite steps for the test suite.
func FeatureContext(s *godog.Suite, bundleLocation *string, pullSecretFile *string) {
	trayHandler = tray.NewTray(bundleLocation, pullSecretFile)
	installerHandler = installer.NewInstaller(&currentUserPassword, &installerPath)
	notificationHandler = notification.NewNotification()
	if handlersAreInitialized() {
		s.Step(`^(.*) CRC app from installer$`,
			executeActionFromInstaller)
		s.Step(`^.*CRC app (.*) installed$`,
			crcInstalled)
		s.Step(`^reboot is required$`,
			installerHandler.RebootRequired)
		s.Step(`^fresh CRC app installation$`,
			guaranteeFreshInstallation)
		// On windows onboarding is shown as first process after reboot
		// on macos this step should include click on app icon
		// both scenearios for openshift preset this step includes set pullsecret
		s.Step(`^onboarding CRC app setting (.*) preset$`,
			trayHandler.Onboarding)
		s.Step(`^CRC app should be running$`,
			trayHandler.IsRunning)
		s.Step(`^CRC app should be accessible$`,
			trayHandler.IsAccessible)
		// todo change this to inspect selected value from config window?
		s.Step(`^CRC app should be ready to start a environment for (.*) preset$`,
			cmd.IsPresetConfig)
		s.Step(`^click (.*) button from app$`,
			clickTrayButton)
		s.Step(`^get notification about the (.*) process$`,
			getNotification)
		s.Step(`^the (.*) instance should be (.*)$`,
			cmd.WaitForInstanceOnState)
		s.Step(`^app shows (.*) as the state for the instance$`,
			checkInstanceStateInApp)
		s.Step(`^a (.*) instance for (.*) preset$`,
			guaranteeInstanceOnState)
		s.Step(`^using copied oc login command for (.*)$`,
			copyOCLoginCommandByUser)
		s.Step(`^user is connected to the cluster as (.*)$`,
			connectClusterAsUser)
		s.BeforeScenario(func(*messages.Pickle) {
			copyRequiredResources(tray.RequiredResourcesPath)
			copyRequiredResources(installer.RequiredResourcesPath)
			copyRequiredResources(notification.RequiredResourcesPath)
		})
	}
}

func ParseFlags() {
	flag.StringVar(&currentUserPassword, "user-password", "",
		"Current user password. User should be sudoer to handle the installation")
	flag.StringVar(&installerPath, "installer-path", "",
		"Full path for installer")
}

func handlersAreInitialized() bool {
	return trayHandler != nil &&
		installerHandler != nil &&
		notificationHandler != nil
}

func copyRequiredResources(requiredResources func() (string, error)) {
	requiredResourcesPath, err := requiredResources()
	if err != nil {
		os.Exit(1)
	}
	if requiredResourcesPath != "" {
		err = util.CopyResourcesFromPath(requiredResourcesPath)
		if err != nil {
			os.Exit(1)
		}
	}
}

func guaranteeFreshInstallation() error {
	err := trayHandler.IsRunning()
	if err != nil {
		return err
	}
	err = trayHandler.IsAccessible()
	if err != nil {
		return err
	}
	err = trayHandler.IsInstanceOnState(tray.InstanceStateStopped)
	if err != nil {
		return err
	}
	err = cmd.CheckMachineNotExists()
	if err != nil {
		return err
	}
	return notificationHandler.ClearNotifications()
}

func guaranteeInstanceOnState(preset, state string) error {
	switch state {
	case "running", "stopped":
		err := util.MatchWithRetry(state, clusterHasState,
			clusterStateWaitRetries, clusterStateWaitTimeout)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s state is not defined as valid cluster state", state)
	}
	return notificationHandler.ClearNotifications()
}

func clusterHasState(expectedState string) error {
	return cmd.CheckCRCStatus(expectedState)
}

func clickTrayButton(action string) error {
	if trayHandler == nil {
		return fmt.Errorf("tray handler should be initialized. Check if tray is supported on your OS")
	}
	switch action {
	case "start":
		return trayHandler.ClickStart()
	case "stop":
		return trayHandler.ClickStop()
	case "delete":
		return trayHandler.ClickDelete()
	case "quit":
		return trayHandler.ClickQuit()
	default:
		return fmt.Errorf("%s action can not be managed through the tray", action)
	}
}

func checkInstanceStateInApp(state string) error {
	if trayHandler == nil {
		return fmt.Errorf("tray handler should be initialized. Check if tray is supported on your OS")
	}
	switch state {
	case "running":
		return trayHandler.IsInstanceOnState(tray.InstanceStateRunning)
	case "stopped":
		return trayHandler.IsInstanceOnState(tray.InstanceStateStopped)
	default:
		return fmt.Errorf("%s state is not defined as valid cluster state", state)
	}
}

func executeActionFromInstaller(action string) error {
	if installerHandler == nil {
		return fmt.Errorf("installer handler should be initialized. Check if installer is supported on your OS")
	}
	switch action {
	case "install":
		return installerHandler.Install()
	default:
		return fmt.Errorf("%s action is not supported by the installer", action)
	}
}

func crcInstalled(state string) error {
	switch state {
	case "is":
		return cmd.CheckCRCExecutableState(cmd.CRCExecutableInstalled)
	case "is not":
		return cmd.CheckCRCExecutableState(cmd.CRCExecutableNotInstalled)
	default:
		return fmt.Errorf("%s state is not defined as valid CRC installation state", state)
	}
}

func getNotification(action string) error {
	if notificationHandler == nil {
		return fmt.Errorf("notification handler should be initialized. Check if tray is supported on your OS")
	}
	switch action {
	case "starting":
		return notificationHandler.CheckProcessNotification(notification.NotificationProcessStart)
	case "stopping":
		return notificationHandler.CheckProcessNotification(notification.NotificationProcessStop)
	case "deleted":
		return notificationHandler.CheckProcessNotification(notification.NotificationProcessDelete)
	default:
		return fmt.Errorf("%s action will not be notified", action)
	}
}

func copyOCLoginCommandByUser(user string) error {
	if trayHandler == nil {
		return fmt.Errorf("tray handler should be initialized. Check if tray is supported on your OS")
	}
	switch user {
	case "kubeadmin":
		return trayHandler.CopyOCLoginCommandAsKubeadmin()
	case "developer":
		return trayHandler.CopyOCLoginCommandAsDeveloper()
	default:
		return fmt.Errorf("tray can not provide login command for user %s", user)
	}
}

func connectClusterAsUser(user string) error {
	if trayHandler == nil {
		return fmt.Errorf("tray handler should be initialized. Check if tray is supported on your OS")
	}
	switch user {
	case "kubeadmin":
		return trayHandler.ConnectClusterAsKubeadmin()
	case "developer":
		return trayHandler.ConnectClusterAsDeveloper()
	default:
		return fmt.Errorf("can not connect to cluster as user %s", user)
	}
}
