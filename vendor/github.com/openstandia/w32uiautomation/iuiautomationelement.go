package w32uiautomation

import (
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
)

type RECT struct {
	Left   uint32
	Top    uint32
	Right  uint32
	Bottom uint32
}

type IUIAutomationElement struct {
	ole.IUnknown
}

type IUIAutomationElementVtbl struct {
	ole.IUnknownVtbl
	SetFocus                        uintptr
	GetRuntimeId                    uintptr
	FindFirst                       uintptr
	FindAll                         uintptr
	FindFirstBuildCache             uintptr
	FindAllBuildCache               uintptr
	BuildUpdatedCache               uintptr
	GetCurrentPropertyValue         uintptr
	GetCurrentPropertyValueEx       uintptr
	GetCachedPropertyValue          uintptr
	GetCachedPropertyValueEx        uintptr
	GetCurrentPatternAs             uintptr
	GetCachedPatternAs              uintptr
	GetCurrentPattern               uintptr
	GetCachedPattern                uintptr
	GetCachedParent                 uintptr
	GetCachedChildren               uintptr
	Get_CurrentProcessId            uintptr
	Get_CurrentControlType          uintptr
	Get_CurrentLocalizedControlType uintptr
	Get_CurrentName                 uintptr
	Get_CurrentAcceleratorKey       uintptr
	Get_CurrentAccessKey            uintptr
	Get_CurrentHasKeyboardFocus     uintptr
	Get_CurrentIsKeyboardFocusable  uintptr
	Get_CurrentIsEnabled            uintptr
	Get_CurrentAutomationId         uintptr
	Get_CurrentClassName            uintptr
	Get_CurrentHelpText             uintptr
	Get_CurrentCulture              uintptr
	Get_CurrentIsControlElement     uintptr
	Get_CurrentIsContentElement     uintptr
	Get_CurrentIsPassword           uintptr
	Get_CurrentNativeWindowHandle   uintptr
	Get_CurrentItemType             uintptr
	Get_CurrentIsOffscreen          uintptr
	Get_CurrentOrientation          uintptr
	Get_CurrentFrameworkId          uintptr
	Get_CurrentIsRequiredForForm    uintptr
	Get_CurrentItemStatus           uintptr
	Get_CurrentBoundingRectangle    uintptr
	Get_CurrentLabeledBy            uintptr
	Get_CurrentAriaRole             uintptr
	Get_CurrentAriaProperties       uintptr
	Get_CurrentIsDataValidForForm   uintptr
	Get_CurrentControllerFor        uintptr
	Get_CurrentDescribedBy          uintptr
	Get_CurrentFlowsTo              uintptr
	Get_CurrentProviderDescription  uintptr
	Get_CachedProcessId             uintptr
	Get_CachedControlType           uintptr
	Get_CachedLocalizedControlType  uintptr
	Get_CachedName                  uintptr
	Get_CachedAcceleratorKey        uintptr
	Get_CachedAccessKey             uintptr
	Get_CachedHasKeyboardFocus      uintptr
	Get_CachedIsKeyboardFocusable   uintptr
	Get_CachedIsEnabled             uintptr
	Get_CachedAutomationId          uintptr
	Get_CachedClassName             uintptr
	Get_CachedHelpText              uintptr
	Get_CachedCulture               uintptr
	Get_CachedIsControlElement      uintptr
	Get_CachedIsContentElement      uintptr
	Get_CachedIsPassword            uintptr
	Get_CachedNativeWindowHandle    uintptr
	Get_CachedItemType              uintptr
	Get_CachedIsOffscreen           uintptr
	Get_CachedOrientation           uintptr
	Get_CachedFrameworkId           uintptr
	Get_CachedIsRequiredForForm     uintptr
	Get_CachedItemStatus            uintptr
	Get_CachedBoundingRectangle     uintptr
	Get_CachedLabeledBy             uintptr
	Get_CachedAriaRole              uintptr
	Get_CachedAriaProperties        uintptr
	Get_CachedIsDataValidForForm    uintptr
	Get_CachedControllerFor         uintptr
	Get_CachedDescribedBy           uintptr
	Get_CachedFlowsTo               uintptr
	Get_CachedProviderDescription   uintptr
	GetClickablePoint               uintptr
}

var IID_IUIAutomationElement = &ole.GUID{0xd22108aa, 0x8ac5, 0x49a5, [8]byte{0x83, 0x7b, 0x37, 0xbb, 0xb3, 0xd7, 0x59, 0x1e}}

func (elem *IUIAutomationElement) VTable() *IUIAutomationElementVtbl {
	return (*IUIAutomationElementVtbl)(unsafe.Pointer(elem.RawVTable))
}

func (elem *IUIAutomationElement) SetFocus() (err error) {
	return setFocus(elem)
}

func (elem *IUIAutomationElement) FindAll(scope TreeScope, condition *IUIAutomationCondition) (found *IUIAutomationElementArray, err error) {
	return findAll(elem, scope, condition)
}

func (elem *IUIAutomationElement) FindFirst(scope TreeScope, condition *IUIAutomationCondition) (found *IUIAutomationElement, err error) {
	return findFirst(elem, scope, condition)
}

func (elem *IUIAutomationElement) GetCurrentPattern(patternId PATTERNID) (*ole.IUnknown, error) {
	return getCurrentPattern(elem, patternId)
}

func (elem *IUIAutomationElement) Get_CurrentAutomationId() (string, error) {
	return get_CurrentAutomationId(elem)
}

func (elem *IUIAutomationElement) Get_CurrentCurrentClassName() (string, error) {
	return get_CurrentClassName(elem)
}

func (elem *IUIAutomationElement) Get_CurrentName() (string, error) {
	return get_CurrentName(elem)
}

func (elem *IUIAutomationElement) Get_CurrentNativeWindowHandle() (syscall.Handle, error) {
	return get_CurrentNativeWindowHandle(elem)
}

func (elem *IUIAutomationElement) Get_CurrentBoundingRectangle() (RECT, error) {
	return get_CurrentBoundingRectangle(elem)
}

func (elem *IUIAutomationElement) Get_CurrentPropertyValue(propertyId PROPERTYID) (ole.VARIANT, error) {
	return get_CurrentPropertyValue(elem, propertyId)
}

func setFocus(elem *IUIAutomationElement) (err error) {
	hr, _, _ := syscall.Syscall(
		elem.VTable().SetFocus,
		1,
		uintptr(unsafe.Pointer(elem)),
		0,
		0)
	if hr != 0 {
		err = ole.NewError(hr)
		return
	}
	return
}

func findAll(elem *IUIAutomationElement, scope TreeScope, condition *IUIAutomationCondition) (found *IUIAutomationElementArray, err error) {
	hr, _, _ := syscall.Syscall6(
		elem.VTable().FindAll,
		4,
		uintptr(unsafe.Pointer(elem)),
		uintptr(scope),
		uintptr(unsafe.Pointer(condition)),
		uintptr(unsafe.Pointer(&found)),
		0,
		0)
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return
}

func findFirst(elem *IUIAutomationElement, scope TreeScope, condition *IUIAutomationCondition) (found *IUIAutomationElement, err error) {
	hr, _, _ := syscall.Syscall6(
		elem.VTable().FindFirst,
		4,
		uintptr(unsafe.Pointer(elem)),
		uintptr(scope),
		uintptr(unsafe.Pointer(condition)),
		uintptr(unsafe.Pointer(&found)),
		0,
		0)
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return
}

func getCurrentPattern(elem *IUIAutomationElement, patternId PATTERNID) (pattern *ole.IUnknown, err error) {
	hr, _, _ := syscall.Syscall(
		elem.VTable().GetCurrentPattern,
		3,
		uintptr(unsafe.Pointer(elem)),
		uintptr(patternId),
		uintptr(unsafe.Pointer(&pattern)))
	if hr != 0 {
		err = ole.NewError(hr)
		return
	}
	return
}

func get_CurrentAutomationId(elem *IUIAutomationElement) (id string, err error) {
	var bstrAutomationId *uint16
	hr, _, _ := syscall.Syscall(
		elem.VTable().Get_CurrentAutomationId,
		2,
		uintptr(unsafe.Pointer(elem)),
		uintptr(unsafe.Pointer(&bstrAutomationId)),
		0)
	if hr != 0 {
		err = ole.NewError(hr)
		return
	}
	id = ole.BstrToString(bstrAutomationId)
	return
}

func get_CurrentClassName(elem *IUIAutomationElement) (name string, err error) {
	var bstrName *uint16
	hr, _, _ := syscall.Syscall(
		elem.VTable().Get_CurrentClassName,
		2,
		uintptr(unsafe.Pointer(elem)),
		uintptr(unsafe.Pointer(&bstrName)),
		0)
	if hr != 0 {
		err = ole.NewError(hr)
		return
	}
	name = ole.BstrToString(bstrName)
	return
}

func get_CurrentName(elem *IUIAutomationElement) (name string, err error) {
	var bstrName *uint16
	hr, _, _ := syscall.Syscall(
		elem.VTable().Get_CurrentName,
		2,
		uintptr(unsafe.Pointer(elem)),
		uintptr(unsafe.Pointer(&bstrName)),
		0)
	if hr != 0 {
		err = ole.NewError(hr)
		return
	}
	name = ole.BstrToString(bstrName)
	return
}

func get_CurrentNativeWindowHandle(elem *IUIAutomationElement) (handle syscall.Handle, err error) {
	hr, _, _ := syscall.Syscall(
		elem.VTable().Get_CurrentNativeWindowHandle,
		2,
		uintptr(unsafe.Pointer(elem)),
		uintptr(unsafe.Pointer(&handle)),
		0)
	if hr != 0 {
		err = ole.NewError(hr)
		return
	}
	return
}

func get_CurrentBoundingRectangle(elem *IUIAutomationElement) (rect RECT, err error) {
	hr, _, _ := syscall.Syscall(
		elem.VTable().Get_CurrentBoundingRectangle,
		2,
		uintptr(unsafe.Pointer(elem)),
		uintptr(unsafe.Pointer(&rect)),
		0)
	if hr != 0 {
		err = ole.NewError(hr)
		return
	}
	return
}

func get_CurrentPropertyValue(elem *IUIAutomationElement, propertyid PROPERTYID) (ole.VARIANT, error) {
	var v ole.VARIANT

	ole.VariantInit(&v)

	hr, _, _ := syscall.Syscall(
		elem.VTable().GetCurrentPropertyValue,
		3,
		uintptr(unsafe.Pointer(elem)),
		uintptr(propertyid),
		uintptr(unsafe.Pointer(&v)))

	if hr != 0 {
		err := ole.NewError(hr)
		return v, err
	}

	return v, nil
}
