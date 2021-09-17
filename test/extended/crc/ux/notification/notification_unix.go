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

func (a applescriptHandler) GetClusterRunning() error {
	return util.MatchWithRetry(startMessage, existNotification,
		notificationWaitRetries, notificationWaitTimeout)
}

func (a applescriptHandler) GetClusterStopped() error {
	return util.MatchWithRetry(stopMessage, existNotification,
		notificationWaitRetries, notificationWaitTimeout)
}

func (a applescriptHandler) GetClusterDeleted() error {
	return util.MatchWithRetry(deleteMessage, existNotification,
		notificationWaitRetries, notificationWaitTimeout)
}

func (a applescriptHandler) ClearNotifications() error {
	return applescript.ExecuteApplescript(manageNotifications, manageNotificationActionClear)
}

func existNotification(expectedNotification string) error {
	return applescript.ExecuteApplescriptReturnShouldMatch(
		expectedNotification, manageNotifications, manageNotificationActionGet)
}
