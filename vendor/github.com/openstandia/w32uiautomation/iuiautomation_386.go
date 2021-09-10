// +build 386
package w32uiautomation

import (
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
)

func createPropertyCondition(aut *IUIAutomation, propertyId PROPERTYID, value ole.VARIANT) (newCondition *IUIAutomationCondition, err error) {
	v := VariantToUintptrArray(value)
	hr, _, _ := syscall.Syscall9(
		aut.VTable().CreatePropertyCondition,
		7,
		uintptr(unsafe.Pointer(aut)),
		uintptr(propertyId),
		v[0],
		v[1],
		v[2],
		v[3],
		uintptr(unsafe.Pointer(&newCondition)),
		0,
		0)
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return
}
