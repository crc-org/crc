package w32uiautomation

type TreeScope uintptr

const (
	TreeScope_Element     = 0x1
	TreeScope_Children    = 0x2
	TreeScope_Descendants = 0x4
	TreeScope_Parent      = 0x8
	TreeScope_Ancestors   = 0x10
	TreeScope_Subtree     = TreeScope_Element | TreeScope_Children | TreeScope_Descendants
)
