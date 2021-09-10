// +build windows
package system_services

import (
	"syscall"

	"github.com/RedHatQE/gowinx/pkg/win32/api/util"
)

var (
	openProcess = kernel32.MustFindProc("OpenProcess")
)

// https://docs.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-openprocess
// HANDLE OpenProcess(
// 	DWORD dwDesiredAccess,
// 	BOOL  bInheritHandle,
// 	DWORD dwProcessId
// );
func OpenProcess(dwDesiredAccess, bInheritHandle uint32, dwProcessId uint32) (handle syscall.Handle, err error) {
	r0, _, e1 := syscall.Syscall(openProcess.Addr(), 3,
		uintptr(dwDesiredAccess),
		uintptr(bInheritHandle),
		uintptr(dwProcessId))
	handle, err = util.EvalSyscallHandler(r0, e1)
	return
}

func OpenProcessAllAccess(inheritHandle bool, processId uint32) (handle syscall.Handle, err error) {
	var inherit uint32 = 0
	if inheritHandle {
		inherit = 1
	}
	return OpenProcess(PROCESS_ALL_ACCESS, inherit, processId)
}
