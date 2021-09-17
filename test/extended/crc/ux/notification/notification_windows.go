// +build windows

package notification

import (
	"fmt"

	actionCenter "github.com/RedHatQE/gowinx/pkg/app/action-center"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/test/extended/util"
)

type gowinxHandler struct {
}

const (
	notificationGroupName string = "CodeReady Containers"
)

func NewNotification() Notification {
	return gowinxHandler{}
}

func RequiredResourcesPath() (string, error) {
	return "", nil
}

func (g gowinxHandler) GetClusterRunning() error {
	return util.MatchWithRetry(startMessage, existNotification,
		notificationWaitRetries, notificationWaitTimeout)
}

func (g gowinxHandler) GetClusterStopped() error {
	return util.MatchWithRetry(stopMessage, existNotification,
		notificationWaitRetries, notificationWaitTimeout)
}

func (g gowinxHandler) GetClusterDeleted() error {
	return util.MatchWithRetry(deleteMessage, existNotification,
		notificationWaitRetries, notificationWaitTimeout)
}

func (g gowinxHandler) ClearNotifications() error {
	if err := actionCenter.ClearNotifications(); err != nil {
		logging.Error(err)
	}
	if err := actionCenter.ClickNotifyButton(); err != nil {
		logging.Error(err)
	}
	return nil
}

func existNotification(expectedNotification string) error {
	notifications, err := actionCenter.GetNotifications(notificationGroupName)
	if err == nil {
		for _, notification := range notifications {
			if notification == expectedNotification {
				return nil
			}
		}
	}
	return fmt.Errorf("%s not found with error", expectedNotification)
}
