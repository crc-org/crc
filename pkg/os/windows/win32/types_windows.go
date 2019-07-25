package win32

type (
	HANDLE uintptr
	HWND   HANDLE
)

const (
	HWND_DESKTOP = HWND(0)
)
