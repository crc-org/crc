package win32

import (
	"syscall"
	"unsafe"

	"github.com/code-ready/crc/pkg/crc/errors"
)

var (
	shell32Lib       = syscall.NewLazyDLL("shell32.dll")
	procShellExecute = shell32Lib.NewProc("ShellExecuteW")
)

// Uses "runas" as verb to execute as Elevated privileges
func ShellExecuteAsAdmin(hwnd HWND, file, parameters, directory string, showCmd int) error {
	return ShellExecute(hwnd, "runas", file, parameters, directory, showCmd)
}

func toUintptr(input string) uintptr {
	return uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(input)))
}

func ShellExecute(hwnd HWND, verb, file, parameters, directory string, showCmd int) error {
	var op, params, dir uintptr
	if len(verb) != 0 {
		op = toUintptr(verb)
	}
	if len(parameters) != 0 {
		params = toUintptr(parameters)
	}
	if len(directory) != 0 {
		dir = toUintptr(directory)
	}

	ret, _, _ := procShellExecute.Call(uintptr(hwnd), op, toUintptr(file), params, dir, uintptr(showCmd))

	if ret == 0 {
		return nil
	}

	return errors.Newf("win32 error %v", ret)
}
