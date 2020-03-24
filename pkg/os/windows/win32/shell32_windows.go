package win32

import (
	"syscall"

	"github.com/code-ready/crc/pkg/crc/logging"
	"golang.org/x/sys/windows"
)

const (
	HWND_DESKTOP = windows.Handle(0)
)

// Uses "runas" as verb to execute as Elevated privileges
func ShellExecuteAsAdmin(reason string, hwnd windows.Handle, file, parameters, directory string, showCmd int) error {
	logging.Infof("Will run as admin: %s", reason)
	return ShellExecute(hwnd, "runas", file, parameters, directory, showCmd)
}

func toUint16ptr(input string) *uint16 {
	uint16ptr, err := syscall.UTF16PtrFromString(input)
	if err != nil {
		logging.Warnf("Failed to convert %s to UTF16: %v", input, err)
	}

	return uint16ptr
}

func ShellExecute(hwnd windows.Handle, verb, file, parameters, directory string, showCmd int) error {
	var op, params, dir *uint16
	if len(verb) != 0 {
		op = toUint16ptr(verb)
	}
	if len(parameters) != 0 {
		params = toUint16ptr(parameters)
	}
	if len(directory) != 0 {
		dir = toUint16ptr(directory)
	}
	return windows.ShellExecute(hwnd, op, toUint16ptr(file), params, dir, int32(showCmd))
}
