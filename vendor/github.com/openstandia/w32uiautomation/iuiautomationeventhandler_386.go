package w32uiautomation

import (
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

type IUIAutomationEventHandler struct {
	ole.IUnknown
	ref int32
}

func automationEventHandler_queryInterface(this *ole.IUnknown, iid *ole.GUID, punk **ole.IUnknown) uint32 {
	*punk = nil
	if ole.IsEqualGUID(iid, ole.IID_IUnknown) ||
		ole.IsEqualGUID(iid, ole.IID_IDispatch) {
		automationEventHandler_addRef(this)
		*punk = this
		return ole.S_OK
	}
	if ole.IsEqualGUID(iid, IID_IUIAutomationEventHandler) {
		automationEventHandler_addRef(this)
		*punk = this
		return ole.S_OK
	}
	return ole.E_NOINTERFACE
}

func automationEventHandler_addRef(this *ole.IUnknown) int32 {
	pthis := (*IUIAutomationEventHandler)(unsafe.Pointer(this))
	pthis.ref++
	return pthis.ref
}

func automationEventHandler_release(this *ole.IUnknown) int32 {
	pthis := (*IUIAutomationEventHandler)(unsafe.Pointer(this))
	pthis.ref--
	return pthis.ref
}
