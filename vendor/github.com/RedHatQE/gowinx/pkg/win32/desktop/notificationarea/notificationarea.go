// +build windows
package notificationarea

import (
	"fmt"
	"syscall"

	"github.com/RedHatQE/gowinx/pkg/util/logging"
	win32wam "github.com/RedHatQE/gowinx/pkg/win32/api/user-interface/windows-and-messages"
	win32toolbar "github.com/RedHatQE/gowinx/pkg/win32/ux/toolbar"
	win32windows "github.com/RedHatQE/gowinx/pkg/win32/ux/windows"
)

// systemtray aka notification area, it is composed of notifications icons (offering display the status and various functions)
// distributerd across:
// * visible area on the right side of the taskbar (class: Shell_TrayWnd)
// * hidden area as overflowwindow ( class: NotifyIconOverflowWindow)

const (
	NOTIFICATION_AREA_VISIBLE_WINDOW_CLASS string = "Shell_TrayWnd"
	NOTIFICATION_AREA_HIDDEN_WINDOW_CLASS  string = "NotifyIconOverflowWindow"
	TOOLBARWINDOWS32_ID                    int32  = 1504
	TRAY_BUTTON_CLASS                      string = "TrayButton"
	ACTION_CENTER_NAME                     string = "Action Center"
)

func GetHiddenNotificationAreaRect() (rect win32wam.RECT, err error) {
	// Show notification area (hidden)
	if err = ShowHiddenNotificationArea(); err == nil {
		if toolbarHandler, err := getNotificationAreaToolbarByWindowClass(NOTIFICATION_AREA_HIDDEN_WINDOW_CLASS); err == nil {
			if _, err = win32wam.GetWindowRect(toolbarHandler, &rect); err == nil {
				logging.Debugf("Rect for system tray t:%d,l:%d,r:%d,b:%d", rect.Top, rect.Left, rect.Right, rect.Bottom)
			}
		}
	}
	if err != nil {
		logging.Errorf("error getting hidden notification area rect: %v\n", err)
	}
	return
}

func ShowHiddenNotificationArea() (err error) {
	if handler, err := getNotificationAreaWindowByClass(NOTIFICATION_AREA_HIDDEN_WINDOW_CLASS); err == nil {
		win32wam.ShowWindow(handler, win32wam.SW_SHOWNORMAL)
	}
	return
}

func getNotificationAreaWindowByClass(className string) (handler syscall.Handle, err error) {
	if handler, err = win32windows.FindWindowByClass(className); err != nil {
		logging.Errorf("error getting handler on notification area for windows class: %s, error: %v\n", className, err)
	}
	return
}

func getNotificationAreaToolbarByWindowClass(className string) (handler syscall.Handle, err error) {
	if windowHandler, err := getNotificationAreaWindowByClass(className); err == nil {
		if handler, err = win32wam.GetDlgItem(windowHandler, TOOLBARWINDOWS32_ID); err != nil {
			logging.Errorf("error getting toolbar handler on notification area for windows class: %s, error: %v\n", className, err)
		}
	}
	return
}

func GetIconPositionByTitle(buttonText string) (int, int, error) {
	toolbarHandlers, _ := findVisibleToolbars()
	for i, toolbarHandler := range toolbarHandlers {
		logging.Debugf("Trying on toolbar %d", i)
		if x, y, err := win32toolbar.GetButtonClickablePosition(toolbarHandler,
			win32toolbar.TOOLBAR_TYPE_VISIBLE,
			buttonText); err == nil {
			return x, y, nil
		}
	}
	toolbarHandler, err := getNotificationAreaToolbarByWindowClass(NOTIFICATION_AREA_HIDDEN_WINDOW_CLASS)
	if err == nil {
		if x, y, err := win32toolbar.GetButtonClickablePosition(toolbarHandler,
			win32toolbar.TOOLBAR_TYPE_HIDDEN,
			buttonText); err == nil {
			return x, y, nil
		}
	}
	return -1, -1, fmt.Errorf("button %s not found on toolbar\n", buttonText)

}

// The notification area is composed of elements, app notification icons use to be placed
// at the toolbars
func findVisibleToolbars() ([]syscall.Handle, error) {
	handler, _ := win32windows.FindWindowByClass(NOTIFICATION_AREA_VISIBLE_WINDOW_CLASS)
	toolbars, _ := win32toolbar.FindToolbars(handler)
	return toolbars, nil
}

func FindTrayButtonByTitle(title string) (syscall.Handle, error) {
	handler, _ := win32windows.FindWindowByClass(NOTIFICATION_AREA_VISIBLE_WINDOW_CLASS)
	return win32windows.FindChildWindowsbyClassAndTitle(handler, TRAY_BUTTON_CLASS, title)
}
