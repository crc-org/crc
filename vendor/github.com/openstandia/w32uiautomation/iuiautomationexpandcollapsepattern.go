package w32uiautomation

import (
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
)

type IUIAutomationExpandCollapsePattern struct {
	ole.IUnknown
}

type IUIAutomationExpandCollapsePatternVtbl struct {
	ole.IUnknownVtbl
	Expand                         uintptr
	Collapse                       uintptr
	Get_CurrentExpandCollapseState uintptr
	Get_CachedExpandCollapseState  uintptr
}

var IID_IUIAutomationExpandCollapsePattern = &ole.GUID{0x619be086, 0x1f4e, 0x4ee4, [8]byte{0xba, 0xfa, 0x21, 0x01, 0x28, 0x73, 0x87, 0x30}}

func (pat *IUIAutomationExpandCollapsePattern) VTable() *IUIAutomationExpandCollapsePatternVtbl {
	return (*IUIAutomationExpandCollapsePatternVtbl)(unsafe.Pointer(pat.RawVTable))
}

func (pat *IUIAutomationExpandCollapsePattern) Expand() error {
	return expand(pat)
}

func (pat *IUIAutomationExpandCollapsePattern) Collapse() error {
	return collapse(pat)
}

func expand(pat *IUIAutomationExpandCollapsePattern) error {
	hr, _, _ := syscall.Syscall(
		pat.VTable().Expand,
		1,
		uintptr(unsafe.Pointer(pat)),
		0,
		0)
	if hr != 0 {
		return ole.NewError(hr)
	}
	return nil
}

func collapse(pat *IUIAutomationExpandCollapsePattern) error {
	hr, _, _ := syscall.Syscall(
		pat.VTable().Collapse,
		1,
		uintptr(unsafe.Pointer(pat)),
		0,
		0)
	if hr != 0 {
		return ole.NewError(hr)
	}
	return nil
}
