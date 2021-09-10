// +build windows
package process

import (
	"syscall"

	win32ss "github.com/RedHatQE/gowinx/pkg/win32/api/system-services"
	win32wam "github.com/RedHatQE/gowinx/pkg/win32/api/user-interface/windows-and-messages"
)

const MEM_COMMIT = 0x1000
const PAGE_READWRITE = 0x04

// Get a process handler for the process holding the window (ux element, represented by its handler)
// This is required in order to run communications with the window.
func GetProcessHandler(windowHandler syscall.Handle) (processHandler syscall.Handle, err error) {
	var tbProcessID uint32
	windowCreatorThreadId, err := win32wam.GetWindowThreadProcessId(windowHandler, &tbProcessID)
	if windowCreatorThreadId > 0 {
		processHandler, err = win32ss.OpenProcessAllAccess(false, tbProcessID)
	}
	return
}

func CloseProcessHandler(processHandler syscall.Handle) error {
	if success, err := win32ss.CloseHandle(processHandler); !success {
		return err
	}
	return nil
}

func AllocateMemory(processHandler syscall.Handle, size int) (uintptr, error) {
	return win32ss.VirtualAllocEx(processHandler, 0, uintptr(size), MEM_COMMIT, PAGE_READWRITE)
}

func FreeMemory(processHandler syscall.Handle, lpBaseAddress uintptr) (bool, error) {
	return win32ss.VirtualFreeEx(processHandler, lpBaseAddress, 0, MEM_COMMIT)
}
