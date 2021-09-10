// +build windows

package ux

import (
	"fmt"

	"github.com/RedHatQE/gowinx/pkg/util/logging"
	win32waf "github.com/RedHatQE/gowinx/pkg/win32/api/user-interface/windows-accesibility-features"
	"github.com/RedHatQE/gowinx/pkg/win32/interaction"
	wa "github.com/openstandia/w32uiautomation"
)

const (
	WINDOW   = "window"
	BUTTON   = "button"
	LIST     = "list"
	LISTITEM = "listitem"
	GROUP    = "group"
	TEXT     = "text"
	MENU     = "menu"
	MENUITEM = "menuitem"
	CHECKBOX = "checkbox"
	EDIT     = "edit"
	PANE     = "pane"
	COMBOBOX = "combobox"

	windowId   = wa.UIA_WindowControlTypeId
	buttonId   = wa.UIA_ButtonControlTypeId
	listId     = wa.UIA_ListControlTypeId
	listitemId = wa.UIA_ListItemControlTypeId
	groupId    = wa.UIA_GroupControlTypeId
	textId     = wa.UIA_TextControlTypeId
	menuId     = wa.UIA_MenuControlTypeId
	menuitemId = wa.UIA_MenuItemControlTypeId
	checkboxId = wa.UIA_CheckBoxControlTypeId
	editId     = wa.UIA_EditControlTypeId
	paneId     = wa.UIA_PaneControlTypeId
	comboboxId = wa.UIA_ComboBoxControlTypeId
)

var elementTypes map[string]int64 = map[string]int64{
	WINDOW:   windowId,
	BUTTON:   buttonId,
	LIST:     listId,
	LISTITEM: listitemId,
	GROUP:    groupId,
	TEXT:     textId,
	MENU:     menuId,
	MENUITEM: menuitemId,
	CHECKBOX: checkboxId,
	EDIT:     editId,
	PANE:     paneId,
	COMBOBOX: comboboxId}

type UXElement struct {
	name        string
	elementType string
	ref         interface{}
}

func Initialize() {
	// Initialize context
	win32waf.Initalize()
}

func Finalize() {
	// Finalize context
	win32waf.Finalize()
}

func GetActiveElement(name string, elementType string) (*UXElement, error) {
	logging.Debugf("Get %s: %s", elementType, name)
	if elementTypeId, ok := elementTypes[elementType]; !ok {
		return nil, fmt.Errorf("Error elementType %s is not supported", elementType)
	} else {
		if element, err := win32waf.GetActiveElement(name, elementTypeId); err != nil || element == nil {
			return nil, fmt.Errorf("Error getting element %s, with error %v", name, err)
		} else {
			return &UXElement{
				name:        name,
				elementType: elementType,
				ref:         element}, nil
		}
	}
}

func (u UXElement) GetName() string {
	return u.name
}

func (u UXElement) GetFullName() string {
	return fmt.Sprintf("%s: %s", u.elementType, u.name)
}

func (u UXElement) Click() error {
	logging.Debug("Click on %s", u.GetFullName())
	if u.elementType == CHECKBOX {
		return wa.Invoke(u.ref.(*wa.IUIAutomationElement))
	} else {
		position, err := win32waf.GetElementRect(u.ref.(*wa.IUIAutomationElement))
		if err != nil {
			return err
		}
		return interaction.ClickOnRect(*position)
	}
}

func (u UXElement) SetValue(value string) error {
	logging.Debug("setting value %s on %s", value, u.GetFullName())
	if u.elementType != EDIT {
		return fmt.Errorf("Error elementType %s is not supported", u.elementType)
	}
	return win32waf.SetElementValue(u.ref.(*wa.IUIAutomationElement), value)

}

func (u UXElement) GetElement(name string, elementType string) (*UXElement, error) {
	if elementTypeId, ok := elementTypes[elementType]; !ok {
		return nil, fmt.Errorf("Error elementType %s is not supported", elementType)
	} else {
		if element, err := win32waf.GetElementFromParent(u.ref.(*wa.IUIAutomationElement), name, elementTypeId); err != nil || element == nil {
			return nil, fmt.Errorf("%s not found on parent %s", elementType, u.GetFullName())
		} else {
			logging.Debugf("Get first %s on parent %s", elementType, u.GetFullName())
			return &UXElement{
				name:        name,
				elementType: elementType,
				ref:         element}, nil
		}
	}
}

func (u UXElement) GetElementByType(elementType string) (*UXElement, error) {
	if elementTypeId, ok := elementTypes[elementType]; !ok {
		return nil, fmt.Errorf("Error elementType %s is not supported", elementType)
	} else {
		if element, err := win32waf.GetElementFromParentByType(u.ref.(*wa.IUIAutomationElement), elementTypeId); err != nil || element == nil {
			return nil, fmt.Errorf("%s not found on parent %s", elementType, u.GetFullName())
		} else {
			name, err := element.Get_CurrentName()
			if err != nil {
				logging.Error(err)
			}
			logging.Debugf("Get first %s on parent %s", elementType, u.GetFullName())
			return &UXElement{
				name:        name,
				elementType: elementType,
				ref:         element}, nil
		}
	}
}

func (u UXElement) GetAllChildren(elementType string) ([]*UXElement, error) {
	logging.Debugf("Get all %s on parent %s ", elementType, u.name)
	if elementTypeId, ok := elementTypes[elementType]; !ok {
		return nil, fmt.Errorf("Error elementType %s is not supported", elementType)
	} else {
		var children []*UXElement
		elements, err := win32waf.GetAllChildren(u.ref.(*wa.IUIAutomationElement), elementTypeId)
		if err != nil {
			return nil, fmt.Errorf("Error getting %s on parent %s with error %v", elementType, u.GetFullName(), err)
		}
		childrenCount, err := elements.Get_Length()
		if err != nil {
			return nil, fmt.Errorf("Error getting %s on parent %s with error %v", elementType, u.GetFullName(), err)
		}
		var i int32
		for i = 0; i < childrenCount; i++ {
			if element, err := elements.GetElement(i); err == nil {
				children = append(children, &UXElement{
					elementType: elementType,
					ref:         element})
			}
		}
		return children, nil
	}
}
