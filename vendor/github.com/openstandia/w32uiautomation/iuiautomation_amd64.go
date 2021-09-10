// +build amd64
package w32uiautomation

import (
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

func createPropertyCondition(aut *IUIAutomation, propertyId PROPERTYID, value ole.VARIANT) (*IUIAutomationCondition, error) {
	var newCondition *IUIAutomationCondition
	hr, _, _ := syscall.Syscall6(
		aut.VTable().CreatePropertyCondition,
		4,
		uintptr(unsafe.Pointer(aut)),
		uintptr(propertyId),
		uintptr(unsafe.Pointer(&value)),
		uintptr(unsafe.Pointer(&newCondition)), 0, 0)
	if hr != 0 {
		return nil, ole.NewError(hr)
	}
	return newCondition, nil
}
