package dns

import (
	"fmt"
	"os"
	"syscall"
)

func sysProcForBackgroundProcess() *syscall.SysProcAttr {
	sysProcAttr := new(syscall.SysProcAttr)
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms682425(v=vs.85).aspx
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms684863(v=vs.85).aspx
	// 0x00000010 = CREATE_NEW_CONSOLE
	sysProcAttr.CreationFlags = syscall.CREATE_NEW_PROCESS_GROUP | 0x00000010
	sysProcAttr.HideWindow = true

	return sysProcAttr
}

func envForBackgroundProcess() []string {
	return []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("PATHEXT=%s", os.Getenv("PATHEXT")),
		fmt.Sprintf("SystemRoot=%s", os.Getenv("SystemRoot")),
		fmt.Sprintf("COMPUTERNAME=%s", os.Getenv("COMPUTERNAME")),
		fmt.Sprintf("TMP=%s", os.Getenv("TMP")),
		fmt.Sprintf("TEMP=%s", os.Getenv("TEMP")),
	}
}
