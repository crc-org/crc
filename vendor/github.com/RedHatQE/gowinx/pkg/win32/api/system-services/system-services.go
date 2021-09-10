// +build windows
package system_services

import (
	"syscall"
)

var (
	kernel32 = syscall.MustLoadDLL("kernel32.dll")
)
