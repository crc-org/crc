// +build windows
package windows_and_messages

// https://docs.microsoft.com/en-us/windows/win32/api/windef/ns-windef-rect
// typedef struct tagRECT {
// 	LONG left;
// 	LONG top;
// 	LONG right;
// 	LONG bottom;
// } RECT, *PRECT, *NPRECT, *LPRECT;
type RECT struct {
	Left, Top, Right, Bottom int32
}

// SendMessage
const (
	WM_USER           uint32 = 1024
	TB_BUTTONCOUNT    uint32 = WM_USER + 24
	TB_COMMANDTOINDEX uint32 = WM_USER + 25
	TB_GETBUTTONINFOW uint32 = WM_USER + 63
	TB_GETBUTTONINFOA uint32 = WM_USER + 65
	TB_GETBUTTONTEXT  uint32 = WM_USER + 75
)

// GetSystemMetrics constants
const (
	SM_CXSCREEN = 0
	SM_CYSCREEN = 1
)

// INPUT Type
const (
	INPUT_MOUSE = 0
)

// MOUSEINPUT DwFlags
const (
	MOUSEEVENTF_ABSOLUTE        = 0x8000
	MOUSEEVENTF_MOVE            = 0x0001
	MOUSEEVENTF_MOVE_NOCOALESCE = 0x2000
	MOUSEEVENTF_LEFTDOWN        = 0x0002
	MOUSEEVENTF_LEFTUP          = 0x0004
)

// Process
const (
	PROCESS_ALL_ACCESS = 0x1F0FFF
)

// ShowWindow constants
const (
	SW_HIDE            = 0
	SW_NORMAL          = 1
	SW_SHOWNORMAL      = 1
	SW_SHOWMINIMIZED   = 2
	SW_MAXIMIZE        = 3
	SW_SHOWMAXIMIZED   = 3
	SW_SHOWNOACTIVATE  = 4
	SW_SHOW            = 5
	SW_MINIMIZE        = 6
	SW_SHOWMINNOACTIVE = 7
	SW_SHOWNA          = 8
	SW_RESTORE         = 9
	SW_SHOWDEFAULT     = 10
	SW_FORCEMINIMIZE   = 11
)
