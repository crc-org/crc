package w32uiautomation

import (
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

func WaitFindFirst(auto *IUIAutomation, elem *IUIAutomationElement, scope TreeScope, condition *IUIAutomationCondition) (found *IUIAutomationElement, err error) {
	for {
		found, err = elem.FindFirst(scope, condition)
		if err != nil {
			return nil, err
		}
		if found != nil {
			return found, nil
		}

		waitChildAdded(auto, elem)
	}
}

func WaitFindAll(auto *IUIAutomation, elem *IUIAutomationElement, scope TreeScope, condition *IUIAutomationCondition) (found *IUIAutomationElementArray, err error) {
	for {
		found, err = elem.FindAll(scope, condition)
		if err != nil {
			return nil, err
		}
		if found != nil {
			return found, nil
		}

		waitChildAdded(auto, elem)
	}
}

func waitChildAdded(auto *IUIAutomation, elem *IUIAutomationElement) error {
	waiting := true
	handler := NewStructureChangedEventHandler(nil)
	lpVtbl := (*IUIAutomationStructureChangedEventHandlerVtbl)(unsafe.Pointer(handler.IUnknown.RawVTable))
	lpVtbl.HandleStructureChangedEvent = syscall.NewCallback(func(this *IUIAutomationStructureChangedEventHandler, sender *IUIAutomationElement, changeType StructureChangeType, runtimeId *ole.SAFEARRAY) syscall.Handle {
		switch changeType {
		case StructureChangeType_ChildAdded, StructureChangeType_ChildrenBulkAdded:
			waiting = false
		}
		return ole.S_OK
	})
	err := auto.AddStructureChangedEventHandler(elem, TreeScope_Subtree, nil, &handler)
	if err != nil {
		return err
	}
	var m ole.Msg
	for waiting {
		ole.GetMessage(&m, 0, 0, 0)
		ole.DispatchMessage(&m)
	}
	err = auto.RemoveStructureChangedEventHandler(elem, &handler)
	if err != nil {
		return err
	}
	return nil
}

