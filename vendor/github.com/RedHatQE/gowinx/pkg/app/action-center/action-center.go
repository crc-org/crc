// +build windows

package action_center

import (
	"fmt"
	"syscall"
	"time"

	"github.com/RedHatQE/gowinx/pkg/util/logging"
	win32wam "github.com/RedHatQE/gowinx/pkg/win32/api/user-interface/windows-and-messages"
	"github.com/RedHatQE/gowinx/pkg/win32/desktop/notificationarea"
	"github.com/RedHatQE/gowinx/pkg/win32/interaction"
	"github.com/RedHatQE/gowinx/pkg/win32/ux"
)

const (
	icon_name        string = "Action Center"
	window_title     string = "Action center"
	clear_all_button string = "Clear all notifications"
)

func ClickNotifyButton() error {
	handler, err := notificationarea.FindTrayButtonByTitle(icon_name)
	if err != nil {
		return err
	}
	rect, err := getActionCenterIconPosition(handler)
	if err != nil {
		return err
	}
	return interaction.ClickOnRect(rect)
}

func ClearNotifications() error {
	// Initialize base elements
	initialize()
	// Get action center window
	actionCenterWindow, err := ux.GetActiveElement(window_title, ux.WINDOW)
	if err != nil {
		logging.Error(err)
		return err
	}
	logging.Info("Got active window for action center")
	clearAllButton, err := actionCenterWindow.GetElement(clear_all_button, ux.BUTTON)
	if err != nil {
		logging.Error(err)
		return err
	}
	if err := clearAllButton.Click(); err != nil {
		return err
	}
	finalize()
	return nil
}

func PrintNotifications(notificationGroupName string) error {
	if notifications, err := GetNotifications(notificationGroupName); err != nil {
		return err
	} else {
		for _, notification := range notifications {
			fmt.Printf("%s", notification)
		}
	}
	return nil
}

// Codeready Containers
// Fix this for testing
func GetNotifications(notificationGroupName string) ([]string, error) {
	notificationsGroup := "Notifications from " + notificationGroupName
	// Initialize base elements
	initialize()
	var notifications []string
	// Get action center window
	actionCenterWindow, err := ux.GetActiveElement(window_title, ux.WINDOW)
	if err != nil {
		logging.Error(err)
		return nil, err
	}
	// Get list of groups of notifications
	// listGroup, err := list.GetList(actionCenterWindow, "")
	listGroup, err := actionCenterWindow.GetElement("", ux.LIST)
	if err != nil {
		logging.Error(err)
		return nil, err
	}

	// Get group of notifications
	// group, err := group.GetGroup(listGroup, notificationsGroup)
	group, err := listGroup.GetElement(notificationsGroup, ux.GROUP)
	if err != nil {
		return nil, err
	}

	// Get notifications on the group
	listItems, err := group.GetAllChildren(ux.LISTITEM)
	if err != nil {
		return nil, err
	}
	for _, listItem := range listItems {
		if textElement, err := listItem.GetElementByType(ux.TEXT); err == nil {
			logging.Infof("Adding notification: %s", textElement.GetName())
			notifications = append(notifications, textElement.GetName())
		}
	}
	finalize()
	return notifications, nil
}

func getActionCenterIconPosition(handler syscall.Handle) (win32wam.RECT, error) {
	var rect win32wam.RECT
	if succeed, err := win32wam.GetWindowRect(handler, &rect); succeed {
		logging.Debugf("Rect for action center icon is t:%d,l:%d,r:%d,b:%d", rect.Top, rect.Left, rect.Right, rect.Bottom)
		return rect, nil
	} else {
		return win32wam.RECT{}, err
	}
}

func initialize() (err error) {
	// Initialize context
	ux.Initialize()
	// Click notifiy button to expand action center
	if err = ClickNotifyButton(); err == nil {
		//delay
		time.Sleep(1 * time.Second)
	}
	return
}

func finalize() {
	// Finalize context
	ux.Finalize()
}
