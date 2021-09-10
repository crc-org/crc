package w32uiautomation

type PATTERNID uintptr

const (
	UIA_AnnotationPatternId        = 10023 // Identifies the Annotation control pattern. Supported starting with Windows 8.
	UIA_DockPatternId              = 10011 // Identifies the Dock control pattern.
	UIA_DragPatternId              = 10030 // Identifies the Drag control pattern. Supported starting with Windows 8.
	UIA_DropTargetPatternId        = 10031 // Identifies the DropTarget control pattern. Supported starting with Windows 8.
	UIA_ExpandCollapsePatternId    = 10005 // Identifies the ExpandCollapse control pattern.
	UIA_GridItemPatternId          = 10007 // Identifies the GridItem control pattern.
	UIA_GridPatternId              = 10006 // Identifies the Grid control pattern.
	UIA_InvokePatternId            = 10000 // Identifies the Invoke control pattern.
	UIA_ItemContainerPatternId     = 10019 // Identifies the ItemContainer control pattern.
	UIA_LegacyIAccessiblePatternId = 10018 // Identifies the LegacyIAccessible control pattern.
	UIA_MultipleViewPatternId      = 10008 // Identifies the MultipleView control pattern.
	UIA_ObjectModelPatternId       = 10022 // Identifies the ObjectModel control pattern. Supported starting with Windows 8.
	UIA_RangeValuePatternId        = 10003 // Identifies the RangeValue control pattern.
	UIA_ScrollItemPatternId        = 10017 // Identifies the ScrollItem control pattern.
	UIA_ScrollPatternId            = 10004 // Identifies the Scroll control pattern.
	UIA_SelectionItemPatternId     = 10010 // Identifies the SelectionItem control pattern.
	UIA_SelectionPatternId         = 10001 // Identifies the Selection control pattern.
	UIA_SpreadsheetPatternId       = 10026 // Identifies the Spreadsheet control pattern. Supported starting with Windows 8.
	UIA_SpreadsheetItemPatternId   = 10027 // Identifies the SpreadsheetItem control pattern. Supported starting with Windows 8.
	UIA_StylesPatternId            = 10025 // Identifies the Styles control pattern. Supported starting with Windows 8.
	UIA_SynchronizedInputPatternId = 10021 // Identifies the SynchronizedInput control pattern.
	UIA_TableItemPatternId         = 10013 // Identifies the TableItem control pattern.
	UIA_TablePatternId             = 10012 // Identifies the Table control pattern.
	UIA_TextChildPatternId         = 10029 // Identifies the TextChild control pattern. Supported starting with Windows 8.
	UIA_TextEditPatternId          = 10032 // Identifies the TextEdit control pattern. Supported starting with Windows 8.1.
	UIA_TextPatternId              = 10014 // Identifies the Text control pattern.
	UIA_TextPattern2Id             = 10024 // Identifies the second version of the Text control pattern. Supported starting with Windows 8.
	UIA_TogglePatternId            = 10015 // Identifies the Toggle control pattern.
	UIA_TransformPatternId         = 10016 // Identifies the Transform control pattern.
	UIA_TransformPattern2Id        = 10028 // Identifies the second version of the Transform control pattern. Supported starting with Windows 8.
	UIA_ValuePatternId             = 10002 // Identifies the Value control pattern.
	UIA_VirtualizedItemPatternId   = 10020 // Identifies the VirtualizedItem control pattern.
	UIA_WindowPatternId            = 10009 // Identifies the Window control pattern.
)
