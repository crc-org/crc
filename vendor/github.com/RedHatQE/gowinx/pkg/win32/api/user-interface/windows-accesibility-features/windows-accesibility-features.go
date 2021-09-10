// +build windows

package windows_accesibility_features

import (
	"syscall"
)

var (
	uiautomationClient = syscall.MustLoadDLL("UIAutomationClient.dll")
)
