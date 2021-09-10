// +build windows

package windows_accesibility_features

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/RedHatQE/gowinx/pkg/util/logging"
	win32wam "github.com/RedHatQE/gowinx/pkg/win32/api/user-interface/windows-and-messages"
	"github.com/go-ole/go-ole"
	wa "github.com/openstandia/w32uiautomation"
)

var (
	manager *wa.IUIAutomation
)

func Initalize() {
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	if waManager, err := wa.NewUIAutomation(); err != nil {
		logging.Errorf("Error initializing ui automation framework: %v", err)
		os.Exit(1)
	} else {
		manager = waManager
	}
}

func Finalize() {
	ole.CoUninitialize()
}

func GetActiveElement(name string, elementType int64) (*wa.IUIAutomationElement, error) {
	root, err := getRootElement()
	defer root.Release()
	if err != nil {
		return nil, err
	} else {
		return GetElementFromParent(root, name, elementType)
	}
}

func GetElementFromParent(parentElement *wa.IUIAutomationElement, name string, elementType int64) (*wa.IUIAutomationElement, error) {
	conditionByName, err := createPropertyCondition(
		wa.UIA_NamePropertyId,
		wa.NewVariantString(name))
	if err != nil {
		return nil, err
	}
	conditionByType, err := createPropertyCondition(
		wa.UIA_ControlTypePropertyId,
		ole.NewVariant(ole.VT_INT, elementType))
	if err != nil {
		return nil, err
	}
	condition, err := createAndCondition(conditionByName, conditionByType)
	if err != nil {
		return nil, err
	}
	// With wait
	// return findFirst(parentElement, wa.TreeScope_Children, condition)
	return parentElement.FindFirst(wa.TreeScope_Children, condition)
}

func GetElementFromParentByType(parentElement *wa.IUIAutomationElement, elementType int64) (*wa.IUIAutomationElement, error) {
	condition, err := createPropertyCondition(
		wa.UIA_ControlTypePropertyId,
		ole.NewVariant(ole.VT_INT, elementType))
	if err != nil {
		return nil, err
	}
	// With wait
	// return findFirst(parentElement, wa.TreeScope_Children, condition)
	return parentElement.FindFirst(wa.TreeScope_Children, condition)
}

func GetAllChildren(parentElement *wa.IUIAutomationElement, elementType int64) (*wa.IUIAutomationElementArray, error) {
	condition, err := createPropertyCondition(
		wa.UIA_ControlTypePropertyId,
		ole.NewVariant(ole.VT_INT, elementType))
	if err != nil {
		return nil, err
	}
	// With wait
	// return findAll(parentElement, wa.TreeScope_Children, condition)
	return parentElement.FindAll(wa.TreeScope_Children, condition)
}

func GetElementRect(element *wa.IUIAutomationElement) (*win32wam.RECT, error) {
	rect, err := element.Get_CurrentBoundingRectangle()
	if err != nil {
		return nil, err
	}
	return &win32wam.RECT{Top: int32(rect.Top),
		Right:  int32(rect.Right),
		Bottom: int32(rect.Bottom),
		Left:   int32(rect.Left)}, nil
}

// // https://github.com/goldendict/goldendict/blob/master/guids.c
// var IID_IUIAutomationTextPattern = &ole.GUID{0x32eba289, 0x3583, 0x42c9, [8]byte{0x9c, 0x59, 0x3b, 0x6d, 0x9a, 0x1e, 0x9b, 0x6a}}
func GetElementText(element *wa.IUIAutomationElement) (string, error) {
	pattern, err := getValuePattern(element)
	if err != nil {
		return "", err
	}
	defer pattern.Release()
	return pattern.Get_CurrentValue()
}

func SetElementValue(element *wa.IUIAutomationElement, value string) error {
	pattern, err := getLegacyIAccessiblePattern(element)
	if err != nil {
		return err
	}
	defer pattern.Release()
	return pattern.SetValue(value)
}

// https://docs.microsoft.com/en-us/windows/win32/api/uiautomationclient/nf-uiautomationclient-iuiautomation-createpropertycondition
// HRESULT CreatePropertyCondition(
// 	PROPERTYID             propertyId,
// 	VARIANT                value,
// 	IUIAutomationCondition **newCondition
//  );
func createPropertyCondition(propertyId wa.PROPERTYID, value ole.VARIANT) (*wa.IUIAutomationCondition, error) {
	var newCondition *wa.IUIAutomationCondition
	hr, _, er1 := syscall.Syscall6(
		manager.VTable().CreatePropertyCondition,
		4,
		uintptr(unsafe.Pointer(manager)),
		uintptr(propertyId),
		uintptr(unsafe.Pointer(&value)),
		uintptr(unsafe.Pointer(&newCondition)),
		0,
		0)
	// https://docs.microsoft.com/en-us/windows/win32/seccrypto/common-hresult-values
	if hr != 0 {
		return nil, error(er1)
	}
	return newCondition, nil
}

func createAndCondition(condition1, condition2 *wa.IUIAutomationCondition) (newCondition *wa.IUIAutomationCondition, err error) {
	return manager.CreateAndCondition(condition1, condition2)
}

// https://docs.microsoft.com/en-us/windows/win32/api/uiautomationclient/nf-uiautomationclient-iuiautomationelement-findfirst
// HRESULT FindFirst(
// 	TreeScope              scope,
// 	IUIAutomationCondition *condition,
// 	IUIAutomationElement   **found
// );
func findFirst(elem *wa.IUIAutomationElement, scope wa.TreeScope, condition *wa.IUIAutomationCondition) (found *wa.IUIAutomationElement, err error) {
	return wa.WaitFindFirst(manager, elem, scope, condition)
}

// https://docs.microsoft.com/en-us/windows/win32/api/uiautomationclient/nf-uiautomationclient-iuiautomationelement-findfirst
// HRESULT FindFirst(
// 	TreeScope              scope,
// 	IUIAutomationCondition *condition,
// 	IUIAutomationElement   **found
// );
func findAll(elem *wa.IUIAutomationElement, scope wa.TreeScope, condition *wa.IUIAutomationCondition) (found *wa.IUIAutomationElementArray, err error) {
	return wa.WaitFindAll(manager, elem, scope, condition)
}

func getRootElement() (root *wa.IUIAutomationElement, err error) {
	return manager.GetRootElement()
}

func getValuePattern(element *wa.IUIAutomationElement) (*wa.IUIAutomationValuePattern, error) {
	unknown, err := element.GetCurrentPattern(wa.UIA_ValuePatternId)
	if err != nil {
		logging.Error(err)
		return nil, err
	}
	logging.Info("found value pattern for element %v", unknown)
	defer unknown.Release()

	disp, err := unknown.QueryInterface(wa.IID_IUIAutomationValuePattern)
	if err != nil {
		logging.Error(err)
		return nil, err
	}
	logging.Info("found interface dispatcher for value pattern")
	return (*wa.IUIAutomationValuePattern)(unsafe.Pointer(disp)), nil
}

func getLegacyIAccessiblePattern(element *wa.IUIAutomationElement) (*IUIAutomationLegacyIAccessiblePattern, error) {
	unknown, err := element.GetCurrentPattern(wa.UIA_LegacyIAccessiblePatternId)
	if err != nil {
		logging.Error(err)
		return nil, err
	}
	logging.Info("found LegacyIAccessible pattern for element %v", unknown)
	defer unknown.Release()

	disp, err := unknown.QueryInterface(IID_IUIAutomationLegacyIAccessiblePattern)
	if err != nil {
		logging.Error(err)
		return nil, err
	}
	logging.Info("found interface dispatcher for LegacyIAccessible pattern")
	return (*IUIAutomationLegacyIAccessiblePattern)(unsafe.Pointer(disp)), nil
}
