package w32uiautomation

import (
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

type IUIAutomationStructureChangedEventHandler struct {
	ole.IUnknown
	ref int64
}

func structureChangedEventHandler_queryInterface(this *ole.IUnknown, iid *ole.GUID, punk **ole.IUnknown) uint64 {
	*punk = nil
	if ole.IsEqualGUID(iid, ole.IID_IUnknown) ||
		ole.IsEqualGUID(iid, ole.IID_IDispatch) {
		structureChangedEventHandler_addRef(this)
		*punk = this
		return ole.S_OK
	}
	if ole.IsEqualGUID(iid, IID_IUIAutomationStructureChangedEventHandler) {
		structureChangedEventHandler_addRef(this)
		*punk = this
		return ole.S_OK
	}
	return ole.E_NOINTERFACE
}

func structureChangedEventHandler_addRef(this *ole.IUnknown) int64 {
	pthis := (*IUIAutomationStructureChangedEventHandler)(unsafe.Pointer(this))
	pthis.ref++
	return pthis.ref
}

func structureChangedEventHandler_release(this *ole.IUnknown) int64 {
	pthis := (*IUIAutomationStructureChangedEventHandler)(unsafe.Pointer(this))
	pthis.ref--
	return pthis.ref
}
