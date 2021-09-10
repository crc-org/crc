package w32uiautomation

import (
	"unsafe"

	"github.com/go-ole/go-ole"
)

type IUIAutomationCacheRequest struct {
	ole.IUnknown
}

type IUIAutomationCacheRequestVtbl struct {
	ole.IUnknownVtbl
}

var IID_IUIAutomationCacheRequest = &ole.GUID{0xb32a92b5, 0xbc25, 0x4078, [8]byte{0x9c, 0x08, 0xd7, 0xee, 0x95, 0xc4, 0x8e, 0x03}}

func (elem *IUIAutomationCacheRequest) VTable() *IUIAutomationCacheRequestVtbl {
	return (*IUIAutomationCacheRequestVtbl)(unsafe.Pointer(elem.RawVTable))
}
