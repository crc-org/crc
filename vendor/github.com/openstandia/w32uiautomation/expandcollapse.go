package w32uiautomation

import "unsafe"

func Expand(element *IUIAutomationElement) error {
	pattern, err := getExpandCollapsePattern(element)
	if err != nil {
		return err
	}
	defer pattern.Release()
	err = pattern.Expand()
	if err != nil {
		return err
	}
	return nil
}

func Collapse(element *IUIAutomationElement) error {
	pattern, err := getExpandCollapsePattern(element)
	if err != nil {
		return err
	}
	defer pattern.Release()
	err = pattern.Collapse()
	if err != nil {
		return err
	}
	return nil
}

func getExpandCollapsePattern(element *IUIAutomationElement) (*IUIAutomationExpandCollapsePattern, error) {
	unknown, err := element.GetCurrentPattern(UIA_ExpandCollapsePatternId)
	if err != nil {
		return nil, err
	}
	defer unknown.Release()

	disp, err := unknown.QueryInterface(IID_IUIAutomationExpandCollapsePattern)
	if err != nil {
		return nil, err
	}

	return (*IUIAutomationExpandCollapsePattern)(unsafe.Pointer(disp)), nil
}
