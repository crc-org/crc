package w32uiautomation

import (
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
)

type IUIAutomationInvokePattern struct {
	ole.IUnknown
}

type IUIAutomationInvokePatternVtbl struct {
	ole.IUnknownVtbl
	Invoke uintptr
}

var IID_IUIAutomationInvokePattern = &ole.GUID{0xfb377fbe, 0x8ea6, 0x46d5, [8]byte{0x9c, 0x73, 0x64, 0x99, 0x64, 0x2d, 0x30, 0x59}}

func (pat *IUIAutomationInvokePattern) VTable() *IUIAutomationInvokePatternVtbl {
	return (*IUIAutomationInvokePatternVtbl)(unsafe.Pointer(pat.RawVTable))
}

func (pat *IUIAutomationInvokePattern) Invoke() error {
	return invoke(pat)
}

func invoke(pat *IUIAutomationInvokePattern) error {
	hr, _, _ := syscall.Syscall(
		pat.VTable().Invoke,
		1,
		uintptr(unsafe.Pointer(pat)),
		0,
		0)
	if hr != 0 {
		return ole.NewError(hr)
	}
	return nil
}
