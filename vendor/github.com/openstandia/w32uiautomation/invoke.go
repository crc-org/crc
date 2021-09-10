package w32uiautomation

import "unsafe"

func Invoke(element *IUIAutomationElement) error {
	unknown, err := element.GetCurrentPattern(UIA_InvokePatternId)
	if err != nil {
		return err
	}
	defer unknown.Release()

	disp, err := unknown.QueryInterface(IID_IUIAutomationInvokePattern)
	if err != nil {
		return err
	}

	pattern := (*IUIAutomationInvokePattern)(unsafe.Pointer(disp))
	defer pattern.Release()
	err = pattern.Invoke()
	if err != nil {
		return err
	}
	return nil
}
