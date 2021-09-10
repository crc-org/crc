package w32uiautomation

import (
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

type IUIAutomationEventHandlerVtbl struct {
	ole.IUnknownVtbl
	HandleAutomationEvent uintptr
}

var IID_IUIAutomationEventHandler = &ole.GUID{0x146c3c17, 0xf12e, 0x4e22, [8]byte{0x8c, 0x27, 0xf8, 0x94, 0xb9, 0xb7, 0x9c, 0x69}}

func (h *IUIAutomationEventHandler) VTable() *IUIAutomationEventHandlerVtbl {
	return (*IUIAutomationEventHandlerVtbl)(unsafe.Pointer(h.RawVTable))
}

func NewAutomationEventHandler(handlerFunc func(this *IUIAutomationEventHandler, sender *IUIAutomationElement, eventId EVENTID) syscall.Handle) IUIAutomationEventHandler {
	lpVtbl := &IUIAutomationEventHandlerVtbl{
		IUnknownVtbl: ole.IUnknownVtbl{
			QueryInterface: syscall.NewCallback(automationEventHandler_queryInterface),
			AddRef:         syscall.NewCallback(automationEventHandler_addRef),
			Release:        syscall.NewCallback(automationEventHandler_release),
		},
		HandleAutomationEvent: syscall.NewCallback(handlerFunc),
	}
	return IUIAutomationEventHandler{
		IUnknown: ole.IUnknown{RawVTable: (*interface{})(unsafe.Pointer(lpVtbl))},
	}
}
