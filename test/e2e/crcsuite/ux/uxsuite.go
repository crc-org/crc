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

// FeatureContext defines godog.Suite steps for the test suite.
func FeatureContext(s *godog.Suite, bundleLocation *string, pullSecretFile *string) {
	trayHandler = tray.NewTray(bundleLocation, pullSecretFile)
	installerHandler = installer.NewInstaller(&currentUserPassword, &installerPath)
	notificationHandler = notification.NewNotification()
	if handlersAreInitialized() {
		s.Step(`^install CRC tray$`,
			trayHandler.Install)
		s.Step(`^tray should be installed$`,
			trayHandler.IsInstalled)
		s.Step(`^tray icon should be accessible$`,
			trayHandler.IsAccessible)
		s.Step(`^fresh tray installation$`,
			guaranteeFreshInstallation)
		s.Step(`^(.*) the cluster from the tray$`,
			clickTrayButton)
		s.Step(`^cluster should be (.*)$`,
			waitForClusterInState)
		s.Step(`^set the pull secret$`,
			trayHandler.SetPullSecret)
		s.Step(`^tray should show cluster as (.*)$`,
			checkClusterStateOnTray)
		s.Step(`^(.*) CRC from installer$`,
			executeActionFromInstaller)
		s.Step(`^.*CRC (.*) installed$`,
			crcInstalled)
		s.Step(`^user should get notified with cluster state as (.*)$`,
			getNotification)
		s.Step(`^a (.*) cluster$`,
			guaranteeClusterState)
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
	err := trayHandler.IsInstalled()
	if err != nil {
		return err
	}
	err = trayHandler.IsAccessible()
	if err != nil {
		return err
	}
	err = trayHandler.IsInitialStatus()
	if err != nil {
		return err
	}
	err = cmd.CheckMachineNotExists()
	if err != nil {
		return err
	}
	return notificationHandler.ClearNotifications()
}

func guaranteeClusterState(state string) error {
	switch state {
	case "running", "stopped":
		err := cmd.CheckCRCStatus(state)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s state is not defined as valid cluster state", state)
	}
	return notificationHandler.ClearNotifications()
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

func checkClusterStateOnTray(state string) error {
	if trayHandler == nil {
		return fmt.Errorf("tray handler should be initialized. Check if tray is supported on your OS")
	}
	switch state {
	case "running":
		return trayHandler.IsClusterRunning()
	case "stopped":
		return trayHandler.IsClusterStopped()
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

func waitForClusterInState(state string) error {
	return cmd.WaitForClusterInState(state)
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
	case "running":
		return notificationHandler.GetClusterRunning()
	case "stopped":
		return notificationHandler.GetClusterStopped()
	case "deleted":
		return notificationHandler.GetClusterDeleted()
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
