// +build windows

package notification

import (
	"fmt"
)

type handler struct {
}

func NewNotification() Notification {
	return handler{}
}

func RequiredResourcesPath() (string, error) {
	return "", nil
}

func (h handler) GetClusterRunning() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) GetClusterStopped() error {
	return fmt.Errorf("not implemented yet")

}

func (h handler) GetClusterDeleted() error {
	return fmt.Errorf("not implemented yet")
}

func (h handler) ClearNotifications() error {
	return fmt.Errorf("not implemented yet")
}
