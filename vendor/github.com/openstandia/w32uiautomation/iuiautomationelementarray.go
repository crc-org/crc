package w32uiautomation
import (
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

type IUIAutomationElementArray struct {
	ole.IUnknown
}

type IUIAutomationElementArrayVtbl struct {
	ole.IUnknownVtbl
	Get_Length uintptr
	GetElement uintptr
}

func (v *IUIAutomationElementArray) VTable() *IUIAutomationElementArrayVtbl {
	return (*IUIAutomationElementArrayVtbl)(unsafe.Pointer(v.RawVTable))
}

func (v *IUIAutomationElementArray) Get_Length() (length int32, err error) {
	hr, _, _ := syscall.Syscall(
		v.VTable().Get_Length,
		2,
		uintptr(unsafe.Pointer(v)),
		uintptr(unsafe.Pointer(&length)),
		0)
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return
}

func (v *IUIAutomationElementArray) GetElement(index int32) (element *IUIAutomationElement, err error) {
	hr, _, _ := syscall.Syscall(
		v.VTable().GetElement,
		3,
		uintptr(unsafe.Pointer(v)),
		uintptr(index),
		uintptr(unsafe.Pointer(&element)))
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return
}

