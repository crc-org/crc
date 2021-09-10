// +build windows

package windows

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/RedHatQE/gowinx/pkg/util/logging"
	win32wam "github.com/RedHatQE/gowinx/pkg/win32/api/user-interface/windows-and-messages"
)

// To get a windows by title among all the windows on the system, it is required
// to loop over all the windows and get the one with the same title text
func FindWindowByTitle(title string) (syscall.Handle, error) {
	var hwnd syscall.Handle
	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		b := make([]uint16, 200)
		_, err := win32wam.GetWindowText(h, &b[0], int32(len(b)))
		if err != nil {
			// ignore the error
			return 1 // continue enumeration
		}
		if syscall.UTF16ToString(b) == title {
			// note the window
			hwnd = h
			return 0 // stop enumeration
		}
		return 1 // continue enumeration
	})
	win32wam.EnumWindows(cb, 0)
	if hwnd == 0 {
		return 0, fmt.Errorf("No window with title '%s' found", title)
	}
	return hwnd, nil
}

func FindWindowByClass(class string) (syscall.Handle, error) {
	z := uint16(0)
	return win32wam.FindWindowW(syscall.StringToUTF16Ptr(class), &z)
}

func FindWindowExByClass(parentHandler syscall.Handle, class string) (syscall.Handle, error) {
	z := uint16(0)
	return win32wam.FindWindowEx(parentHandler, syscall.Handle(0), syscall.StringToUTF16Ptr(class), &z)
}

// Get all child windows whith class
func FindChildWindowsbyClass(hwndParent syscall.Handle, class string) ([]syscall.Handle, error) {
	var hwnds []syscall.Handle
	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		elementClassName, err := win32wam.GetClassName(h)
		if err != nil {
			return 1 // continue enumeration
		}
		if elementClassName == class {
			hwnds = append(hwnds, h)
		}
		return 1 // continue enumeration
	})
	win32wam.EnumChildWindows(hwndParent, cb, 0)
	if len(hwnds) == 0 {
		return hwnds, fmt.Errorf("No child element with classname %s\n", class)
	}
	return hwnds, nil
}

// Get child window with title
func FindChildWindowByTitle(hwndParent syscall.Handle, title string) (syscall.Handle, int32, error) {
	var hwnd syscall.Handle
	var elementIndex int32
	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		b := make([]uint16, 200)
		_, err := win32wam.GetWindowText(h, &b[0], int32(len(b)))
		if err != nil {
			elementIndex++
		}
		elementTitle := syscall.UTF16ToString(b)
		logging.Debugf("looking for child elements got: %s", elementTitle)
		if elementTitle == title {
			hwnd = h
			return 0 // stop enumeration
		}
		elementIndex++
		return 1 // continue enumeration
	})
	win32wam.EnumChildWindows(hwndParent, cb, 0)
	if hwnd == 0 {
		logging.Errorf("Error the expected element with title %s", title)
		return 0, 0, fmt.Errorf("No window with title '%s' found", title)
	}
	return hwnd, elementIndex, nil
}

// Get all child windows whith class
func FindChildWindowsbyClassAndTitle(hwndParent syscall.Handle, class, title string) (syscall.Handle, error) {
	var hwnd syscall.Handle
	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		elementClassName, err := win32wam.GetClassName(h)
		if err != nil {
			return 1 // continue enumeration
		}
		if elementClassName == class {
			b := make([]uint16, 200)
			_, err := win32wam.GetWindowText(h, &b[0], int32(len(b)))
			if err != nil {
				return 1 // continue enumeration
			}
			elementTitle := syscall.UTF16ToString(b)
			logging.Debugf("looking for child elements got: %s", elementTitle)
			if strings.Contains(strings.ToLower(elementTitle), strings.ToLower(title)) {
				hwnd = h
				return 0 // stop enumeration
			}
		}
		return 1 // continue enumeration
	})
	win32wam.EnumChildWindows(hwndParent, cb, 0)
	if hwnd == 0 {
		return 0, fmt.Errorf("No child element with classname %s\n", class)
	}
	return hwnd, nil
}

// Get all child windows whith class
func FindChildren(hwndParent syscall.Handle) ([]syscall.Handle, error) {
	var hwnds []syscall.Handle
	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		elementClassName, err := win32wam.GetClassName(h)
		if err != nil {
			return 1 // continue enumeration
		}
		logging.Debugf("looking for child elements got: %s", elementClassName)
		b := make([]uint16, 200)
		_, err = win32wam.GetWindowText(h, &b[0], int32(len(b)))
		if err != nil {
			return 1 // continue enumeration
		}
		elementTitle := syscall.UTF16ToString(b)
		logging.Debugf("looking for child elements got: %s", elementTitle)
		hwnds = append(hwnds, h)
		return 1 // continue enumeration
	})
	win32wam.EnumChildWindows(hwndParent, cb, 0)
	return hwnds, nil
}
