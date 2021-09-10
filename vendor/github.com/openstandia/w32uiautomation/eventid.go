package w32uiautomation

type EVENTID uintptr

const (
	UIA_ActiveTextPositionChangedEventId                 = 20032 // Identifies the event that is raised when the active text position changes, indicated by a navigation event within or between read-only text elements (such as web browsers, PDF documents, or EPUB documents) using bookmarks (fragment identifiers that refer to a location within a resource).
	UIA_AsyncContentLoadedEventId                        = 20006 // Identifies the event that is raised when asynchronous content is being loaded. This event is used mainly by providers to indicate that asynchronous content-loading events have occurred.
	UIA_AutomationFocusChangedEventId                    = 20005 // Identifies the event that is raised when the focus has changed from one element to another.
	UIA_AutomationPropertyChangedEventId                 = 20004 // Identifies the event that is raised when the value of a property has changed.
	UIA_ChangesEventId                                   = 20034 // Identifies the event that is raised when a provider calls the UiaRaiseChangesEvent function.
	UIA_Drag_DragCancelEventId                           = 20027 // Identifies the event that is raised when the user ends a drag operation before dropping an element on a drop target. This event is raised by the element being dragged. Supported starting with Windows 8.
	UIA_Drag_DragCompleteEventId                         = 20028 // Identifies the event that is raised when the user drops an element on a drop target. This event is raised by the element being dragged. Supported starting with Windows 8.
	UIA_Drag_DragStartEventId                            = 20026 // Identifies the event that is raised when the user starts to drag an element. This event is raised by the element being dragged. Supported starting with Windows 8.
	UIA_DropTarget_DragEnterEventId                      = 20029 // Identifies the event that is raised when the user drags an element into a drop target's boundary. This event is raised by the drop target element. Supported starting with Windows 8.
	UIA_DropTarget_DragLeaveEventId                      = 20030 // Identifies the event that is raised when the user drags an element out of a drop target's boundary. This event is raised by the drop target element. Supported starting with Windows 8.
	UIA_DropTarget_DroppedEventId                        = 20031 // Identifies the event that is raised when the user drops an element on a drop target. This event is raised by the drop target element. Supported starting with Windows 8.
	UIA_HostedFragmentRootsInvalidatedEventId            = 20025 // Identifies the event that is raised when a change is made to the root node of a UI Automation fragment that is hosted in another element. Supported starting with Windows 8.
	UIA_InputDiscardedEventId                            = 20022 // Identifies the event that is raised when the specified input was discarded or otherwise failed to reach any element.
	UIA_InputReachedOtherElementEventId                  = 20021 // Identifies the event that is raised when the specified input reached an element other than the element for which the StartListening method was called.
	UIA_InputReachedTargetEventId                        = 20020 // Identifies the event that is raised when the specified mouse or keyboard input reaches the element for which the StartListening method was called.
	UIA_Invoke_InvokedEventId                            = 20009 // Identifies the event that is raised when a control is invoked or activated.
	UIA_LayoutInvalidatedEventId                         = 20008 // Identifies the event that is raised when the layout of child items within a control has changed. This event is also used for Auto-suggest accessibility.
	UIA_LiveRegionChangedEventId                         = 20024 // Identifies the event that is raised when the content of a live region has changed. Supported starting with Windows 8.
	UIA_MenuClosedEventId                                = 20007 // Identifies the event that is raised when a menu is closed.
	UIA_MenuModeEndEventId                               = 20019 // Identifies the event that is raised when a menu mode is ended.
	UIA_MenuModeStartEventId                             = 20018 // Identifies the event that is raised when a menu mode is started.
	UIA_MenuOpenedEventId                                = 20003 // Identifies the event that is raised when a menu is opened.
	UIA_NotificationEventId                              = 20035 // Identifies the event that is raised when a provider calls the UiaRaiseNotificationEvent method.
	UIA_Selection_InvalidatedEventId                     = 20013 // Identifies the event that is raised when a selection in a container has changed significantly.
	UIA_SelectionItem_ElementAddedToSelectionEventId     = 20010 // Identifies the event raised when an item is added to a collection of selected items.
	UIA_SelectionItem_ElementRemovedFromSelectionEventId = 20011 // Identifies the event raised when an item is removed from a collection of selected items.
	UIA_SelectionItem_ElementSelectedEventId             = 20012 // Identifies the event that is raised when a call to the Select, AddToSelection, or RemoveFromSelection method results in a single item being selected.
	UIA_StructureChangedEventId                          = 20002 // Identifies the event that is raised when the UI Automation tree structure is changed.
	UIA_SystemAlertEventId                               = 20023 // Identifies the event that is raised when a provider issues a system alert. Supported starting with Windows 8.
	UIA_Text_TextChangedEventId                          = 20015 // Identifies the event that is raised whenever textual content is modified.
	UIA_Text_TextSelectionChangedEventId                 = 20014 // Identifies the event that is raised when the text selection is modified.
	UIA_TextEdit_ConversionTargetChangedEventId          = 20033 // Identifies the event that is raised whenever a composition replacement is performed by a control. Supported starting with Windows 8.1.
	UIA_TextEdit_TextChangedEventId                      = 20032 // Identifies the event that is raised whenever text auto-correction is performed by a control. Supported starting with Windows 8.1.
	UIA_ToolTipClosedEventId                             = 20001 // Identifies the event that is raised when a tooltip is closed.
	UIA_ToolTipOpenedEventId                             = 20000 // Identifies the event that is raised when a tooltip is opened.
	UIA_Window_WindowClosedEventId                       = 20017 // Identifies the event that is raised when a window is closed.
	UIA_Window_WindowOpenedEventId                       = 20016 // Identifies the event that is raised when a window is opened.
)
