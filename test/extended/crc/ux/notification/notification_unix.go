//go:build !windows
// +build !windows

package notification

import (
	"runtime"

	"github.com/code-ready/crc/test/extended/os/applescript"
	"github.com/code-ready/crc/test/extended/util"
)

type applescriptHandler struct {
}

const (
	// startMessage  string = "OpenShift cluster is running"
	// stopMessage   string = "The OpenShift Cluster was successfully stopped"
	// deleteMessage string = "The OpenShift Cluster is successfully deleted"

	scriptsRelativePath           string = "applescripts"
	manageNotifications           string = "manageNotifications.applescript"
	manageNotificationActionGet   string = "get"
	manageNotificationActionClear string = "clear"
)

func NewNotification() Notification {
	if runtime.GOOS == "darwin" {
		return applescriptHandler{}
	}
	return nil
}

func RequiredResourcesPath() (string, error) {
	return applescript.GetScriptsPath(scriptsRelativePath)
}

func (a applescriptHandler) CheckProcessNotification(process string) error {
	return util.MatchWithRetry(process, existNotification,
		notificationWaitRetries, notificationWaitTimeout)
}

func (a applescriptHandler) ClearNotifications() error {
	return applescript.ExecuteApplescript(manageNotifications, manageNotificationActionClear)
}

func existNotification(expectedNotification string) error {
	return applescript.ExecuteApplescriptReturnShouldMatch(
		expectedNotification, manageNotifications, manageNotificationActionGet)
}
