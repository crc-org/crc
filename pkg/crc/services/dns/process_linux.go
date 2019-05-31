package dns

import (
	"fmt"
	"os"
	"syscall"
)

func sysProcForBackgroundProcess() *syscall.SysProcAttr {
	sysProcAttr := new(syscall.SysProcAttr)
	sysProcAttr.Setpgid = true
	sysProcAttr.Pgid = 0

	return sysProcAttr
}

func envForBackgroundProcess() []string {
	return []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	}
}
