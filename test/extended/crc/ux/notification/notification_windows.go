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
	startMessage  string = "CodeReady Containers Cluster has started"
	stopMessage   string = "Cluster stopped"
	deleteMessage string = "Cluster deleted"

	notificationGroupName string = "CodeReady Containers"
)

func NewNotification() Notification {
	return gowinxHandler{}
}

func RequiredResourcesPath() (string, error) {
	return "", nil
}

func (g gowinxHandler) GetClusterRunning() error {
	return clusterStateNotified(startMessage)
}

func (g gowinxHandler) GetClusterStopped() error {
	return clusterStateNotified(stopMessage)
}

func (g gowinxHandler) GetClusterDeleted() error {
	return clusterStateNotified(deleteMessage)
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

func clusterStateNotified(notificationMessage string) error {
	err := util.MatchWithRetry(notificationMessage, existNotification,
		notificationWaitRetries, notificationWaitTimeout)
	return err
}

func existNotification(expectedNotification string) error {
	notifications, err := actionCenter.GetNotifications(notificationGroupName)
	if errClear := actionCenter.ClickNotifyButton(); err != nil {
		logging.Error(errClear)
	}
	if errClear := actionCenter.ClearNotifications(); err != nil {
		logging.Error(errClear)
	}
	if errClear := actionCenter.ClickNotifyButton(); err != nil {
		logging.Error(errClear)
	}
	if err == nil {
		for _, notification := range notifications {
			logging.Infof("Expected notification is %s", expectedNotification)
			logging.Infof("Current notification is %s", notification)
			if notification == expectedNotification {
				return nil
			}
		}

	}
	return fmt.Errorf("%s not found with error", expectedNotification)
}
