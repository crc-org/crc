package w32uiautomation

import (
	"unsafe"

	"github.com/go-ole/go-ole"
)

type IUIAutomationCondition struct {
	ole.IUnknown
}

type IUIAutomationConditionVtbl struct {
	ole.IUnknownVtbl
}

func (v *IUIAutomationCondition) VTable() *IUIAutomationConditionVtbl {
	return (*IUIAutomationConditionVtbl)(unsafe.Pointer(v.RawVTable))
}
