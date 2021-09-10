package w32uiautomation

import (
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
)

type IUIAutomationValuePattern struct {
	ole.IUnknown
}

type IUIAutomationValuePatternVtbl struct {
	ole.IUnknownVtbl
	SetValue                   uintptr
	Get_CurrentValue           uintptr
	Get_CurrentIsReadonly      uintptr
	Get_CachedValue            uintptr
	Get_CachedCachedIsReadOnly uintptr
}

var IID_IUIAutomationValuePattern = &ole.GUID{0xa94cd8b1, 0x0844, 0x4cd6, [8]byte{0x9d, 0x2d, 0x64, 0x05, 0x37, 0xab, 0x39, 0xe9}}

func (pat *IUIAutomationValuePattern) VTable() *IUIAutomationValuePatternVtbl {
	return (*IUIAutomationValuePatternVtbl)(unsafe.Pointer(pat.RawVTable))
}

func (pat *IUIAutomationValuePattern) SetValue(value string) error {
	s16, err := syscall.UTF16PtrFromString(value)
	if err != nil {
		return err
	}
	hr, _, _ := syscall.Syscall(
		pat.VTable().SetValue,
		2,
		uintptr(unsafe.Pointer(pat)),
		uintptr(unsafe.Pointer(s16)),
		0)
	if hr != 0 {
		return ole.NewError(hr)
	}
	return nil
}

func (pat *IUIAutomationValuePattern) Get_CurrentValue() (name string, err error) {
	var bstrName *uint16
	hr, _, _ := syscall.Syscall(
		pat.VTable().Get_CurrentValue,
		2,
		uintptr(unsafe.Pointer(pat)),
		uintptr(unsafe.Pointer(&bstrName)),
		0)
	if hr != 0 {
		err = ole.NewError(hr)
		return
	}
	name = ole.BstrToString(bstrName)
	return
}
