// +build windows
package windows_and_messages

import (
	"syscall"
)

var (
	user32 = syscall.MustLoadDLL("user32.dll")
)
