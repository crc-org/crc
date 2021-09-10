package w32uiautomation

import (
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
)

type IUIAutomationSelectionItemPattern struct {
	ole.IUnknown
}

type IUIAutomationSelectionItemPatternVtbl struct {
	ole.IUnknownVtbl
	Select                        uintptr
	AddToSelection                uintptr
	RemoveFromSelection           uintptr
	Get_CurrentIsSelected         uintptr
	Get_CurrentSelectionContainer uintptr
	Get_CachedIsSelected          uintptr
	Get_CachedSelectionContainer  uintptr
}

var IID_IUIAutomationSelectionItemPattern = &ole.GUID{0xa8efa66a, 0x0fda, 0x421a, [8]byte{0x91, 0x94, 0x38, 0x02, 0x1f, 0x35, 0x78, 0xea}}

func (pat *IUIAutomationSelectionItemPattern) VTable() *IUIAutomationSelectionItemPatternVtbl {
	return (*IUIAutomationSelectionItemPatternVtbl)(unsafe.Pointer(pat.RawVTable))
}

func (pat *IUIAutomationSelectionItemPattern) Select() error {
	return select_(pat)
}

func select_(pat *IUIAutomationSelectionItemPattern) error {
	hr, _, _ := syscall.Syscall(
		pat.VTable().Select,
		1,
		uintptr(unsafe.Pointer(pat)),
		0,
		0)
	if hr != 0 {
		return ole.NewError(hr)
	}
	return nil
}
