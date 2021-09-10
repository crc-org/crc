package w32uiautomation

import (
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

type IUIAutomation struct {
	ole.IUnknown
}

type IUIAutomationVtbl struct {
	ole.IUnknownVtbl
	CompareElements                           uintptr
	CompareRuntimeIds                         uintptr
	GetRootElement                            uintptr
	ElementFromHandle                         uintptr
	ElementFromPoint                          uintptr
	GetFocusedElement                         uintptr
	GetRootElementBuildCache                  uintptr
	ElementFromHandleBuildCache               uintptr
	ElementFromPointBuildCache                uintptr
	GetFocusedElementBuildCache               uintptr
	CreateTreeWalker                          uintptr
	Get_ControlViewWalker                     uintptr
	Get_ContentViewWalker                     uintptr
	Get_RawViewWalker                         uintptr
	Get_RawViewCondition                      uintptr
	Get_ControlViewCondition                  uintptr
	Get_ContentViewCondition                  uintptr
	CreateCacheRequest                        uintptr
	CreateTrueCondition                       uintptr
	CreateFalseCondition                      uintptr
	CreatePropertyCondition                   uintptr
	CreatePropertyConditionEx                 uintptr
	CreateAndCondition                        uintptr
	CreateAndConditionFromArray               uintptr
	CreateAndConditionFromNativeArray         uintptr
	CreateOrCondition                         uintptr
	CreateOrConditionFromArray                uintptr
	CreateOrConditionFromNativeArray          uintptr
	CreateNotCondition                        uintptr
	AddAutomationEventHandler                 uintptr
	RemoveAutomationEventHandler              uintptr
	AddPropertyChangedEventHandlerNativeArray uintptr
	AddPropertyChangedEventHandler            uintptr
	RemovePropertyChangedEventHandler         uintptr
	AddStructureChangedEventHandler           uintptr
	RemoveStructureChangedEventHandler        uintptr
	AddFocusChangedEventHandler               uintptr
	RemoveFocusChangedEventHandler            uintptr
	RemoveAllEventHandlers                    uintptr
	IntNativeArrayToSafeArray                 uintptr
	IntSafeArrayToNativeArray                 uintptr
	RectToVariant                             uintptr
	VariantToRect                             uintptr
	SafeArrayToRectNativeArray                uintptr
	CreateProxyFactoryEntry                   uintptr
	Get_ProxyFactoryMapping                   uintptr
	GetPropertyProgrammaticName               uintptr
	GetPatternProgrammaticName                uintptr
	PollForPotentialSupportedPatterns         uintptr
	PollForPotentialSupportedProperties       uintptr
	CheckNotSupported                         uintptr
	Get_ReservedNotSupportedValue             uintptr
	Get_ReservedMixedAttributeValue           uintptr
	ElementFromIAccessible                    uintptr
	ElementFromIAccessibleBuildCache          uintptr
}

var CLSID_CUIAutomation = &ole.GUID{0xff48dba4, 0x60ef, 0x4201, [8]byte{0xaa, 0x87, 0x54, 0x10, 0x3e, 0xef, 0x59, 0x4e}}

var IID_IUIAutomation = &ole.GUID{0x30cbe57d, 0xd9d0, 0x452a, [8]byte{0xab, 0x13, 0x7a, 0xc5, 0xac, 0x48, 0x25, 0xee}}

func (v *IUIAutomation) VTable() *IUIAutomationVtbl {
	return (*IUIAutomationVtbl)(unsafe.Pointer(v.RawVTable))
}

func NewUIAutomation() (*IUIAutomation, error) {
	auto, err := ole.CreateInstance(CLSID_CUIAutomation, IID_IUIAutomation)
	if err != nil {
		return nil, err
	}
	return (*IUIAutomation)(unsafe.Pointer(auto)), nil
}

func (auto *IUIAutomation) CompareElements(el1, el2 *IUIAutomation) (areSame bool, err error) {
	return compareElements(auto, el1, el2)
}

func (auto *IUIAutomation) GetRootElement() (root *IUIAutomationElement, err error) {
	return getRootElement(auto)
}

func (auto *IUIAutomation) CreateTreeWalker(condition *IUIAutomationCondition) (walker *IUIAutomationTreeWalker, err error) {
	return createTreeWalker(auto, condition)
}

func (auto *IUIAutomation) CreateTrueCondition() (newCondition *IUIAutomationCondition, err error) {
	return createTrueCondition(auto)
}

func (auto *IUIAutomation) CreateAndCondition(condition1, condition2 *IUIAutomationCondition) (newCondition *IUIAutomationCondition, err error) {
	return createAndCondition(auto, condition1, condition2)
}

func (auto *IUIAutomation) CreatePropertyCondition(propertyId PROPERTYID, value ole.VARIANT) (newCondition *IUIAutomationCondition, err error) {
	return createPropertyCondition(auto, propertyId, value)
}

func (auto *IUIAutomation) AddAutomationEventHandler(eventId EVENTID, element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationEventHandler) error {
	return addAutomationEventHandler(auto, eventId, element, scope, cacheRequest, handler)
}

func (auto *IUIAutomation) RemoveAutomationEventHandler(eventId EVENTID, element *IUIAutomationElement, handler *IUIAutomationEventHandler) error {
	return removeAutomationEventHandler(auto, eventId, element, handler)
}

func (auto *IUIAutomation) AddStructureChangedEventHandler(element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationStructureChangedEventHandler) error {
	return addStructureChangedEventHandler(auto, element, scope, cacheRequest, handler)
}

func (auto *IUIAutomation) RemoveStructureChangedEventHandler(element *IUIAutomationElement, handler *IUIAutomationStructureChangedEventHandler) error {
	return removeStructureChangedEventHandler(auto, element, handler)
}

func (auto *IUIAutomation) RemoveAllEventHandlers() error {
	return removeAllEventHandlers(auto)
}

func compareElements(auto *IUIAutomation, el1, el2 *IUIAutomation) (areSame bool, err error) {
	hr, _, _ := syscall.Syscall6(
		auto.VTable().CompareElements,
		4,
		uintptr(unsafe.Pointer(auto)),
		uintptr(unsafe.Pointer(el1)),
		uintptr(unsafe.Pointer(el2)),
		uintptr(unsafe.Pointer(&areSame)),
		0,
		0)
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return
}

func getRootElement(auto *IUIAutomation) (root *IUIAutomationElement, err error) {
	hr, _, _ := syscall.Syscall(
		auto.VTable().GetRootElement,
		3,
		uintptr(unsafe.Pointer(auto)),
		uintptr(unsafe.Pointer(&root)),
		0)
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return
}

func createTreeWalker(auto *IUIAutomation, condition *IUIAutomationCondition) (walker *IUIAutomationTreeWalker, err error) {
	hr, _, _ := syscall.Syscall(
		auto.VTable().CreateTreeWalker,
		3,
		uintptr(unsafe.Pointer(auto)),
		uintptr(unsafe.Pointer(condition)),
		uintptr(unsafe.Pointer(&walker)))
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return
}

func createTrueCondition(auto *IUIAutomation) (newCondition *IUIAutomationCondition, err error) {
	hr, _, _ := syscall.Syscall(
		auto.VTable().CreateTrueCondition,
		2,
		uintptr(unsafe.Pointer(auto)),
		uintptr(unsafe.Pointer(&newCondition)),
		0)
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return
}

func createAndCondition(auto *IUIAutomation, condition1, condition2 *IUIAutomationCondition) (newCondition *IUIAutomationCondition, err error) {
	hr, _, _ := syscall.Syscall6(
		auto.VTable().CreateAndCondition,
		4,
		uintptr(unsafe.Pointer(auto)),
		uintptr(unsafe.Pointer(condition1)),
		uintptr(unsafe.Pointer(condition2)),
		uintptr(unsafe.Pointer(&newCondition)),
		0,
		0)
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return
}

func addAutomationEventHandler(auto *IUIAutomation, eventId EVENTID, element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationEventHandler) error {
	hr, _, _ := syscall.Syscall6(
		auto.VTable().AddAutomationEventHandler,
		6,
		uintptr(unsafe.Pointer(auto)),
		uintptr(unsafe.Pointer(eventId)),
		uintptr(unsafe.Pointer(element)),
		uintptr(scope),
		uintptr(unsafe.Pointer(cacheRequest)),
		uintptr(unsafe.Pointer(handler)))
	if hr != 0 {
		return ole.NewError(hr)
	}
	return nil
}

func removeAutomationEventHandler(auto *IUIAutomation, eventId EVENTID, element *IUIAutomationElement, handler *IUIAutomationEventHandler) error {
	hr, _, _ := syscall.Syscall6(
		auto.VTable().RemoveAutomationEventHandler,
		4,
		uintptr(unsafe.Pointer(auto)),
		uintptr(unsafe.Pointer(eventId)),
		uintptr(unsafe.Pointer(element)),
		uintptr(unsafe.Pointer(handler)),
		0,
		0)
	if hr != 0 {
		return ole.NewError(hr)
	}
	return nil
}

func addStructureChangedEventHandler(auto *IUIAutomation, element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationStructureChangedEventHandler) error {
	hr, _, _ := syscall.Syscall6(
		auto.VTable().AddStructureChangedEventHandler,
		5,
		uintptr(unsafe.Pointer(auto)),
		uintptr(unsafe.Pointer(element)),
		uintptr(scope),
		uintptr(unsafe.Pointer(cacheRequest)),
		uintptr(unsafe.Pointer(handler)),
		0)
	if hr != 0 {
		return ole.NewError(hr)
	}
	return nil
}

func removeStructureChangedEventHandler(auto *IUIAutomation, element *IUIAutomationElement, handler *IUIAutomationStructureChangedEventHandler) error {
	hr, _, _ := syscall.Syscall(
		auto.VTable().RemoveStructureChangedEventHandler,
		3,
		uintptr(unsafe.Pointer(auto)),
		uintptr(unsafe.Pointer(element)),
		uintptr(unsafe.Pointer(handler)))
	if hr != 0 {
		return ole.NewError(hr)
	}
	return nil
}

func removeAllEventHandlers(auto *IUIAutomation) error {
	hr, _, _ := syscall.Syscall(
		auto.VTable().RemoveAllEventHandlers,
		1,
		uintptr(unsafe.Pointer(auto)),
		0,
		0)
	if hr != 0 {
		return ole.NewError(hr)
	}
	return nil
}
