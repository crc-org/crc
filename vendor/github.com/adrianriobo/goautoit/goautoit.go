// +build windows
// +build amd64

package goautoit

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

//properties available in AutoItX.
const (
	SWHide            = 0
	SWMaximize        = 3
	SWMinimize        = 6
	SWRestore         = 9
	SWShow            = 5
	SWShowDefault     = 10
	SWShowMaximized   = 3
	SWShowMinimized   = 2
	SWShowminNoActive = 7
	SWShowMa          = 8
	SWShowNoActive    = 4
	SWShowNormal      = 1

	INTDEFAULT         = -2147483647
	DefaultMouseButton = "left"

	libraryFolder = "external"
	autoItX3      = "AutoItX3_x64.dll"
)

// HWND -- window handle
type HWND uintptr

//RECT -- http://msdn.microsoft.com/en-us/library/windows/desktop/dd162897.aspx
type RECT struct {
	Left, Top, Right, Bottom int32
}

//POINT --
type POINT struct {
	X, Y int32
}

var (
	dll64                   *syscall.LazyDLL
	clipGet                 *syscall.LazyProc
	clipPut                 *syscall.LazyProc
	controlClick            *syscall.LazyProc
	controlClickByHandle    *syscall.LazyProc
	controlCommand          *syscall.LazyProc
	controlCommandByHandle  *syscall.LazyProc
	controlDisable          *syscall.LazyProc
	controlDisableByHandle  *syscall.LazyProc
	controlEnable           *syscall.LazyProc
	controlEnableByHandle   *syscall.LazyProc
	controlFocus            *syscall.LazyProc
	controlFocusByHandle    *syscall.LazyProc
	controlGetHandle        *syscall.LazyProc
	controlGetHandleAsText  *syscall.LazyProc
	controlGetPos           *syscall.LazyProc
	controlGetPosByHandle   *syscall.LazyProc
	controlGetText          *syscall.LazyProc
	controlGetTextByHandle  *syscall.LazyProc
	controlHide             *syscall.LazyProc
	controlHideByHandle     *syscall.LazyProc
	controlListView         *syscall.LazyProc
	controlListViewByHandle *syscall.LazyProc
	controlMove             *syscall.LazyProc
	controlMoveByHandle     *syscall.LazyProc
	controlSend             *syscall.LazyProc
	controlSendByHandle     *syscall.LazyProc
	controlSetText          *syscall.LazyProc
	controlSetTextByHandle  *syscall.LazyProc
	controlShow             *syscall.LazyProc
	controlShowByHandle     *syscall.LazyProc
	controlTreeView         *syscall.LazyProc
	controlTreeViewByHandle *syscall.LazyProc
	isAdmin                 *syscall.LazyProc
	mouseClick              *syscall.LazyProc
	mouseClickDrag          *syscall.LazyProc
	mouseDown               *syscall.LazyProc
	mouseGetCursor          *syscall.LazyProc
	mouseGetPos             *syscall.LazyProc
	mouseMove               *syscall.LazyProc
	mouseUp                 *syscall.LazyProc
	mouseWheel              *syscall.LazyProc
	opt                     *syscall.LazyProc
	processClose            *syscall.LazyProc
	processExists           *syscall.LazyProc
	processSetPriority      *syscall.LazyProc
	processWait             *syscall.LazyProc
	processWaitClose        *syscall.LazyProc
	run                     *syscall.LazyProc
	runAs                   *syscall.LazyProc
	runAsWait               *syscall.LazyProc
	runWait                 *syscall.LazyProc
	send                    *syscall.LazyProc
	winActivate             *syscall.LazyProc
	winActive               *syscall.LazyProc
	winCloseByHandle        *syscall.LazyProc
	winGetHandle            *syscall.LazyProc
	winGetText              *syscall.LazyProc
	winGetTitle             *syscall.LazyProc
	winMinimizeAll          *syscall.LazyProc
	winMinimizeAllundo      *syscall.LazyProc
	winMove                 *syscall.LazyProc
	winGetState             *syscall.LazyProc
	winSetState             *syscall.LazyProc
	winWait                 *syscall.LazyProc
)

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	dll64, err := loadLibrary(autoItX3)
	if err != nil {
		os.Exit(1)
	}
	clipGet = dll64.NewProc("AU3_ClipGet")
	clipPut = dll64.NewProc("AU3_ClipPut")
	controlClick = dll64.NewProc("AU3_ControlClick")
	controlClickByHandle = dll64.NewProc("AU3_ControlClickByHandle")
	controlCommand = dll64.NewProc("AU3_ControlCommand")
	controlCommandByHandle = dll64.NewProc("AU3_ControlCommandByHandle")
	controlGetHandle = dll64.NewProc("AU3_ControlGetHandle")
	controlGetHandleAsText = dll64.NewProc("AU3_ControlGetHandleAsText")
	controlGetPos = dll64.NewProc("AU3_ControlGetPos")
	controlGetPosByHandle = dll64.NewProc("AU3_ControlGetPosByHandle")
	controlGetText = dll64.NewProc("AU3_ControlGetText")
	controlGetTextByHandle = dll64.NewProc("AU3_ControlGetTextByHandle")
	controlHide = dll64.NewProc("AU3_ControlHide")
	controlHideByHandle = dll64.NewProc("AU3_ControlHideByHandle")
	controlListView = dll64.NewProc("AU3_ControlListView")
	controlListViewByHandle = dll64.NewProc("AU3_ControlListViewByHandle")
	controlMove = dll64.NewProc("AU3_ControlMove")
	controlMoveByHandle = dll64.NewProc("AU3_ControlMoveByHandle")
	controlSend = dll64.NewProc("AU3_ControlSend")
	controlSendByHandle = dll64.NewProc("AU3_ControlSendByHandle")
	controlSetText = dll64.NewProc("AU3_ControlSetText")
	controlSetTextByHandle = dll64.NewProc("AU3_ControlSetTextByHandle")
	controlShow = dll64.NewProc("AU3_ControlShow")
	controlShowByHandle = dll64.NewProc("AU3_ControlShowByHandle")
	controlTreeView = dll64.NewProc("AU3_ControlTreeView")
	controlTreeViewByHandle = dll64.NewProc("AU3_ControlTreeViewByHandle")
	isAdmin = dll64.NewProc("AU3_IsAdmin")
	mouseClick = dll64.NewProc("AU3_MouseClick")
	mouseClickDrag = dll64.NewProc("AU3_MouseClickDrag")
	mouseDown = dll64.NewProc("AU3_MouseDown")
	mouseGetCursor = dll64.NewProc("AU3_MouseGetCursor")
	mouseGetPos = dll64.NewProc("AU3_MouseGetPos")
	mouseMove = dll64.NewProc("AU3_MouseMove")
	mouseUp = dll64.NewProc("AU3_MouseUp")
	mouseWheel = dll64.NewProc("AU3_MouseWheel")
	opt = dll64.NewProc("AU3_Opt")
	processClose = dll64.NewProc("AU3_ProcessClose")
	processExists = dll64.NewProc("AU3_ProcessExists")
	processSetPriority = dll64.NewProc("AU3_ProcessSetPriority")
	processWait = dll64.NewProc("AU3_ProcessWait")
	processWaitClose = dll64.NewProc("AU3_ProcessWaitClose")
	run = dll64.NewProc("AU3_Run")
	runAs = dll64.NewProc("AU3_RunAs")
	runAsWait = dll64.NewProc("AU3_RunAsWait")
	runWait = dll64.NewProc("AU3_RunWait")
	send = dll64.NewProc("AU3_Send")
	winActivate = dll64.NewProc("AU3_WinActivate")
	winActive = dll64.NewProc("AU3_WinActive")
	winCloseByHandle = dll64.NewProc("AU3_WinCloseByHandle")
	winGetHandle = dll64.NewProc("AU3_WinGetHandle")
	winGetText = dll64.NewProc("AU3_WinGetText")
	winGetTitle = dll64.NewProc("AU3_WinGetTitle")
	winMinimizeAll = dll64.NewProc("AU3_WinMinimizeAll")
	winMinimizeAllundo = dll64.NewProc("AU3_WinMinimizeAllUndo")
	winMove = dll64.NewProc("AU3_WinMove")
	winGetState = dll64.NewProc("AU3_WinGetState")
	winSetState = dll64.NewProc("AU3_WinSetState")
	winWait = dll64.NewProc("AU3_WinWait")
}

func loadLibrary(library string) (*syscall.LazyDLL, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return nil, fmt.Errorf("error loading library %s\n", library)
	}
	libraryPath := string(path.Dir(filename) + string(filepath.Separator) +
		libraryFolder + string(filepath.Separator) + library)
	return syscall.NewLazyDLL(libraryPath), nil
}

// WinMinimizeAll -- all windows should be minimize
func WinMinimizeAll() {
	winMinimizeAll.Call()
}

//WinMinimizeAllUndo -- undo minimize all windows
func WinMinimizeAllUndo() {
	winMinimizeAllundo.Call()
}

//WinGetTitle -- get windows title
func WinGetTitle(szTitle, szText string, bufSize int) string {
	// szTitle := "[active]"
	// szText := ""
	// bufSize := 256
	buff := make([]uint16, int(bufSize))
	ret, _, lastErr := winGetTitle.Call(strPtr(szTitle), strPtr(szText), uintptr(unsafe.Pointer(&buff[0])), intPtr(bufSize))
	log.Println(ret)
	log.Println(lastErr)
	return (goWString(buff))
}

//WinGetText -- get text in window
func WinGetText(szTitle, szText string, bufSize int) string {
	buff := make([]uint16, int(bufSize))
	winGetText.Call(strPtr(szTitle), strPtr(szText), uintptr(unsafe.Pointer(&buff[0])), intPtr(bufSize))
	return (goWString(buff))
}

// Run -- Run a windows program
// flag 3(max) 6(min) 9(normal) 0(hide)
func Run(szProgram string, args ...interface{}) int {
	var szDir string
	var flag int
	var ok bool
	if len(args) == 0 {
		szDir = ""
		flag = SWShowNormal
	} else if len(args) == 1 {
		if szDir, ok = args[0].(string); !ok {
			panic("szDir must be a string")
		}
		flag = SWShowNormal
	} else if len(args) == 2 {
		if szDir, ok = args[0].(string); !ok {
			panic("szDir must be a string")
		}
		if flag, ok = args[1].(int); !ok {
			panic("flag must be a int")
		}
	} else {
		panic("Too more parameter")
	}
	pid, _, lastErr := run.Call(strPtr(szProgram), strPtr(szDir), intPtr(flag))
	// log.Println(pid)
	if int(pid) == 0 {
		log.Println(lastErr)
	}
	return int(pid)
}

//Send -- Send simulates input on the keyboard
// flag: 0: normal, 1: raw
func Send(key string, args ...interface{}) {
	var nMode int
	var ok bool
	if len(args) == 0 {
		nMode = 0
	} else if len(args) == 1 {
		if nMode, ok = args[0].(int); !ok {
			panic("nMode must be a int")
		}
	} else {
		panic("Too more parameter")
	}
	send.Call(strPtr(key), intPtr(nMode))
}

//WinWait -- wait window to active
//
func WinWait(szTitle string, args ...interface{}) int {
	var szText string
	var nTimeout int
	var ok bool
	if len(args) == 0 {
		szText = ""
		nTimeout = 0
	} else if len(args) == 1 {
		if szText, ok = args[0].(string); !ok {
			panic("szText must be a string")
		}
		nTimeout = 0
	} else if len(args) == 2 {
		if szText, ok = args[0].(string); !ok {
			panic("szText must be a string")
		}
		if nTimeout, ok = args[1].(int); !ok {
			panic("nTimeout must be a int")
		}
	} else {
		panic("Too more parameter")
	}

	handle, _, lastErr := winWait.Call(strPtr(szTitle), strPtr(szText), intPtr(nTimeout))
	if int(handle) == 0 {
		log.Print("timeout or failure!!!")
		log.Println(lastErr)
	}
	return int(handle)
}

//MouseClick -- Perform a mouse click operation.
func MouseClick(button string, args ...interface{}) int {
	var x, y, nClicks, nSpeed int
	var ok bool

	if len(args) == 0 {
		x = INTDEFAULT
		y = INTDEFAULT
		nClicks = 1
		nSpeed = 10
	} else if len(args) == 2 {
		if x, ok = args[0].(int); !ok {
			panic("x must be a int")
		}
		if y, ok = args[1].(int); !ok {
			panic("y must be a int")
		}
		nClicks = 1
		nSpeed = 10
	} else if len(args) == 3 {
		if x, ok = args[0].(int); !ok {
			panic("x must be a int")
		}
		if y, ok = args[1].(int); !ok {
			panic("y must be a int")
		}
		if nClicks, ok = args[2].(int); !ok {
			panic("nClicks must be a int")
		}
		nSpeed = 10
	} else if len(args) == 4 {
		if x, ok = args[0].(int); !ok {
			panic("x must be a int")
		}
		if y, ok = args[1].(int); !ok {
			panic("y must be a int")
		}
		if nClicks, ok = args[2].(int); !ok {
			panic("nClicks must be a int")
		}
		if nSpeed, ok = args[3].(int); !ok {
			panic("nSpeed must be a int")
		}
	} else {
		panic("Error parameters")
	}
	ret, _, lastErr := mouseClick.Call(strPtr(button), intPtr(x), intPtr(y), intPtr(nClicks), intPtr(nSpeed))
	if int(ret) != 1 {
		log.Print("failure!!!")
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlClick -- Sends a mouse click command to a given control.
func ControlClick(title, text, control string, args ...interface{}) int {
	var button string
	var x, y, nClicks int
	var ok bool

	if len(args) == 0 {
		button = DefaultMouseButton
		nClicks = 1
		x = INTDEFAULT
		y = INTDEFAULT
	} else if len(args) == 1 {
		if button, ok = args[0].(string); !ok {
			panic("button must be a string")
		}
		nClicks = 1
		x = INTDEFAULT
		y = INTDEFAULT
	} else if len(args) == 2 {
		if button, ok = args[0].(string); !ok {
			panic("button must be a string")
		}
		if nClicks, ok = args[1].(int); !ok {
			panic("nClicks must be a int")
		}
		x = INTDEFAULT
		y = INTDEFAULT
	} else if len(args) == 4 {
		if button, ok = args[0].(string); !ok {
			panic("button must be a string")
		}
		if nClicks, ok = args[1].(int); !ok {
			panic("nClicks must be a int")
		}
		if x, ok = args[2].(int); !ok {
			panic("x must be a int")
		}
		if y, ok = args[3].(int); !ok {
			panic("y must be a int")
		}
	} else {
		panic("Error parameters")
	}
	ret, _, lastErr := controlClick.Call(strPtr(title), strPtr(text), strPtr(control), strPtr(button), intPtr(nClicks), intPtr(x), intPtr(y))
	if int(ret) == 0 {
		log.Print("failure!!!")
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlClickByHandle -- Sends a mouse click command to a given control.
func ControlClickByHandle(handle, control HWND, args ...interface{}) int {
	var button string
	var x, y, nClicks int
	var ok bool

	if len(args) == 0 {
		button = DefaultMouseButton
		nClicks = 1
		x = INTDEFAULT
		y = INTDEFAULT
	} else if len(args) == 1 {
		if button, ok = args[0].(string); !ok {
			panic("button must be a string")
		}
		nClicks = 1
		x = INTDEFAULT
		y = INTDEFAULT
	} else if len(args) == 2 {
		if button, ok = args[0].(string); !ok {
			panic("button must be a string")
		}
		if nClicks, ok = args[1].(int); !ok {
			panic("nClicks must be a int")
		}
		x = INTDEFAULT
		y = INTDEFAULT
	} else if len(args) == 4 {
		if button, ok = args[0].(string); !ok {
			panic("button must be a string")
		}
		if nClicks, ok = args[1].(int); !ok {
			panic("nClicks must be a int")
		}
		if x, ok = args[2].(int); !ok {
			panic("x must be a int")
		}
		if y, ok = args[3].(int); !ok {
			panic("y must be a int")
		}
	} else {
		panic("Error parameters")
	}
	ret, _, lastErr := controlClickByHandle.Call(uintptr(handle), uintptr(control), strPtr(button), intPtr(nClicks), intPtr(x), intPtr(y))
	if int(ret) == 0 {
		log.Print("failure!!!")
		log.Println(lastErr)
	}
	return int(ret)
}

//ClipGet -- get a string from clip
func ClipGet(args ...interface{}) string {
	var nBufSize int
	var ok bool
	if len(args) == 0 {
		nBufSize = 256
	} else if len(args) == 1 {
		if nBufSize, ok = args[0].(int); !ok {
			panic("nBufSize must be a int")
		}
	} else {
		panic("Error parameters")
	}
	clip := make([]uint16, int(nBufSize))
	clipGet.Call(uintptr(unsafe.Pointer(&clip[0])), intPtr(nBufSize))
	return (goWString(clip))
}

// ClipPut -- put a string to clip
func ClipPut(szClip string) int {
	ret, _, lastErr := clipPut.Call(strPtr(szClip))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

// WinActivate ( "title" [, "text"]) int
func WinActivate(title string, args ...interface{}) int {
	text := ""
	var ok bool
	argsLen := len(args)
	if argsLen > 1 {
		panic("argument count > 2")
	}
	if argsLen == 1 {
		if text, ok = args[0].(string); !ok {
			panic("text must be a string")
		}
	}
	ret, _, lastErr := winActivate.Call(strPtr(title), strPtr(text))
	if int(ret) == 0 {
		log.Print("failure!!!")
		log.Println(lastErr)
	}
	return int(ret)
}

// WinActive ( "title" [, "text"]) int
func WinActive(title string, args ...interface{}) int {
	text := ""
	var ok bool
	argsLen := len(args)
	if argsLen > 1 {
		panic("argument count > 2")
	}
	if argsLen == 1 {
		if text, ok = args[0].(string); !ok {
			panic("text must be a string")
		}
	}
	ret, _, lastErr := winActive.Call(strPtr(title), strPtr(text))
	if int(ret) == 0 {
		log.Print("failure!!!")
		log.Println(lastErr)
	}
	return int(ret)
}

// WinGetHandle -- get window handle
func WinGetHandle(title string, args ...interface{}) HWND {
	var text string
	var ok bool
	if len(args) == 0 {
		text = ""
	} else if len(args) == 1 {
		if text, ok = args[0].(string); !ok {
			panic("text must be a string")
		}
	} else {
		panic("Error parameters")
	}
	ret, _, lastErr := winGetHandle.Call(strPtr(title), strPtr(text))
	if int(ret) == 0 {
		log.Print("failure!!!")
		log.Println(lastErr)
	}
	return HWND(ret)
}

// WinMove ( "title", "text", x, y [, width [, height [, speed]]] ) int
func WinMove(title, text string, x, y int, args ...interface{}) int {
	width := INTDEFAULT
	height := INTDEFAULT
	speed := 10
	var ok bool
	argsLen := len(args)
	if argsLen > 0 {
		if width, ok = args[0].(int); !ok {
			panic("width must be an integer")
		}
		if argsLen > 1 {
			if height, ok = args[1].(int); !ok {
				panic("height must be an integer")
			}
			if argsLen > 2 {
				if speed, ok = args[2].(int); !ok {
					panic("speed must be an integer")
				}
				if speed < 1 || speed > 100 {
					panic("speed must in range 1~100(slowest)")
				}
				if argsLen > 3 {
					panic("too many arguments")
				}
			}
		}
	}
	ret, _, lastErr := winMove.Call(strPtr(title), strPtr(text), intPtr(x), intPtr(y), intPtr(width),
		intPtr(height), intPtr(speed))
	if int(ret) == 0 {
		log.Print("failure!!!")
		log.Println(lastErr)
	}
	return int(ret)
}

// WinCloseByHandle --
func WinCloseByHandle(hwnd HWND) int {
	ret, _, lastErr := winCloseByHandle.Call(uintptr(hwnd))
	if int(ret) == 0 {
		log.Print("failure!!!")
		log.Println(lastErr)
	}
	return int(ret)
}

// WinGetState ( "title" [, "text"] ) int
func WinGetState(title string, args ...interface{}) int {
	text := ""
	var ok bool
	argsLen := len(args)
	if argsLen > 1 {
		panic("argument count > 2")
	}
	if argsLen == 1 {
		if text, ok = args[0].(string); !ok {
			panic("text must be a string")
		}
	}
	ret, _, lastErr := winGetState.Call(strPtr(title), strPtr(text))
	if int(ret) == 0 {
		log.Println("winGetState failure!!!", lastErr)
	}
	return int(ret)
}

// WinSetState ( "title", "text", flag) int
func WinSetState(title, text string, flag int) int {
	ret, _, lastErr := winSetState.Call(strPtr(title), strPtr(text), intPtr(flag))
	if int(ret) == 0 {
		log.Print("WinSetState failure!!!")
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlSend -- Sends a string of characters to a control.
func ControlSend(title, text, control, sendText string, args ...interface{}) int {
	var nMode int
	var ok bool
	if len(args) == 0 {
		nMode = 0
	} else if len(args) == 1 {
		if nMode, ok = args[0].(int); !ok {
			panic("nMode must be a int")
		}
	} else {
		panic("Too more parameter")
	}
	ret, _, lastErr := controlSend.Call(strPtr(title), strPtr(text), strPtr(control), strPtr(sendText), intPtr(nMode))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlSendByHandle -- Sends a string of characters to a control.
func ControlSendByHandle(handle, control HWND, sendText string, args ...interface{}) int {
	var nMode int
	var ok bool
	if len(args) == 0 {
		nMode = 0
	} else if len(args) == 1 {
		if nMode, ok = args[0].(int); !ok {
			panic("nMode must be a int")
		}
	} else {
		panic("Too more parameter")
	}
	ret, _, lastErr := controlSendByHandle.Call(uintptr(handle), uintptr(control), strPtr(sendText), intPtr(nMode))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlSetText -- Sets text of a control.
func ControlSetText(title, text, control, newText string) int {
	ret, _, lastErr := controlSetText.Call(strPtr(title), strPtr(text), strPtr(control), strPtr(newText))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlSetTextByHandle -- Sets text of a control.
func ControlSetTextByHandle(handle, control HWND, newText string) int {
	ret, _, lastErr := controlSetTextByHandle.Call(uintptr(handle), uintptr(control), strPtr(newText))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlCommand -- Sends a command to a control.
func ControlCommand(title, text, control, command string, args ...interface{}) string {
	var Extra string
	var bufSize int
	var ok bool

	if len(args) == 0 {
		Extra = ""
		bufSize = 256
	} else if len(args) == 1 {
		if Extra, ok = args[0].(string); !ok {
			panic("Extra must be a string")
		}
		bufSize = 256
	} else if len(args) == 2 {
		if Extra, ok = args[0].(string); !ok {
			panic("Extra must be a string")
		}
		if bufSize, ok = args[1].(int); !ok {
			panic("bufferSize must be a int")
		}
	} else {
		panic("Error parameters")
	}
	buff := make([]uint16, int(bufSize))
	ret, _, lastErr := controlCommand.Call(strPtr(title), strPtr(text), strPtr(control), strPtr(command), strPtr(Extra), uintptr(unsafe.Pointer(&buff[0])), intPtr(bufSize))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return (goWString(buff))
}

//ControlCommandByHandle -- Sends a command to a control.
func ControlCommandByHandle(handle, control HWND, command string, args ...interface{}) string {
	var Extra string
	var bufSize int
	var ok bool

	if len(args) == 0 {
		Extra = ""
		bufSize = 256
	} else if len(args) == 1 {
		if Extra, ok = args[0].(string); !ok {
			panic("Extra must be a string")
		}
		bufSize = 256
	} else if len(args) == 2 {
		if Extra, ok = args[0].(string); !ok {
			panic("Extra must be a string")
		}
		if bufSize, ok = args[1].(int); !ok {
			panic("bufferSize must be a int")
		}
	} else {
		panic("Error parameters")
	}
	buff := make([]uint16, int(bufSize))
	ret, _, lastErr := controlCommandByHandle.Call(uintptr(handle), uintptr(control), strPtr(command), strPtr(Extra), uintptr(unsafe.Pointer(&buff[0])), intPtr(bufSize))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return (goWString(buff))
}

//ControlListView --Sends a command to a ListView32 control.
func ControlListView(title, text, control, command string, args ...interface{}) string {
	var Extra1, Extra2 string
	var bufSize int
	var ok bool

	if len(args) == 0 {
		Extra1 = ""
		Extra2 = ""
		bufSize = 256
	} else if len(args) == 1 {
		if Extra1, ok = args[0].(string); !ok {
			panic("Extra1 must be a string")
		}
		Extra2 = ""
		bufSize = 256
	} else if len(args) == 2 {
		if Extra1, ok = args[0].(string); !ok {
			panic("Extra1 must be a string")
		}
		if Extra2, ok = args[1].(string); !ok {
			panic("Extra2 must be a string")
		}
		bufSize = 256
	} else if len(args) == 3 {
		if Extra1, ok = args[0].(string); !ok {
			panic("Extra1 must be a string")
		}
		if Extra2, ok = args[1].(string); !ok {
			panic("Extra2 must be a string")
		}
		if bufSize, ok = args[2].(int); !ok {
			panic("bufSize must be a int")
		}
	} else {
		panic("Error parameters")
	}
	buff := make([]uint16, int(bufSize))
	ret, _, lastErr := controlListView.Call(strPtr(title), strPtr(text), strPtr(control), strPtr(command), strPtr(Extra1), strPtr(Extra2), uintptr(unsafe.Pointer(&buff[0])), intPtr(bufSize))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return (goWString(buff))
}

//ControlListViewByHandle --Sends a command to a ListView32 control.
func ControlListViewByHandle(handle, control HWND, command string, args ...interface{}) string {
	var Extra1, Extra2 string
	var bufSize int
	var ok bool

	if len(args) == 0 {
		Extra1 = ""
		Extra2 = ""
		bufSize = 256
	} else if len(args) == 1 {
		if Extra1, ok = args[0].(string); !ok {
			panic("Extra1 must be a string")
		}
		Extra2 = ""
		bufSize = 256
	} else if len(args) == 2 {
		if Extra1, ok = args[0].(string); !ok {
			panic("Extra1 must be a string")
		}
		if Extra2, ok = args[1].(string); !ok {
			panic("Extra2 must be a string")
		}
		bufSize = 256
	} else if len(args) == 3 {
		if Extra1, ok = args[0].(string); !ok {
			panic("Extra1 must be a string")
		}
		if Extra2, ok = args[1].(string); !ok {
			panic("Extra2 must be a string")
		}
		if bufSize, ok = args[2].(int); !ok {
			panic("bufSize must be a int")
		}
	} else {
		panic("Error parameters")
	}
	buff := make([]uint16, int(bufSize))
	ret, _, lastErr := controlListViewByHandle.Call(uintptr(handle), uintptr(control), strPtr(command), strPtr(Extra1), strPtr(Extra2), uintptr(unsafe.Pointer(&buff[0])), intPtr(bufSize))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return (goWString(buff))
}

//ControlDisable -- Disables or "grays-out" a control.
func ControlDisable(title, text, control string) int {
	ret, _, lastErr := controlDisable.Call(strPtr(title), strPtr(text), strPtr(control))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlDisableByHandle -- Disables or "grays-out" a control.
func ControlDisableByHandle(handle, control HWND) int {
	ret, _, lastErr := controlDisableByHandle.Call(uintptr(handle), uintptr(control))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlEnable -- Enables a "grayed-out" control.
func ControlEnable(title, text, control string) int {
	ret, _, lastErr := controlEnable.Call(strPtr(title), strPtr(text), strPtr(control))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlEnableByHandle -- Enables a "grayed-out" control.
func ControlEnableByHandle(handle, control HWND) int {
	ret, _, lastErr := controlEnableByHandle.Call(uintptr(handle), uintptr(control))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlFocus -- Sets input focus to a given control on a window.
func ControlFocus(title, text, control string) int {
	ret, _, lastErr := controlFocus.Call(strPtr(title), strPtr(text), strPtr(control))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlFocusByHandle -- Sets input focus to a given control on a window.
func ControlFocusByHandle(handle, control HWND) int {
	ret, _, lastErr := controlFocusByHandle.Call(uintptr(handle), uintptr(control))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlGetHandle -- Retrieves the internal handle of a control.
func ControlGetHandle(handle HWND, control string) HWND {
	ret, _, lastErr := controlGetHandle.Call(uintptr(handle), strPtr(control))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return HWND(ret)
}

//ControlGetHandleAsText -- Retrieves the internal handle of a control.
func ControlGetHandleAsText(title, text, control string, args ...interface{}) string {
	var bufSize int
	var ok bool

	if len(args) == 0 {
		bufSize = 256
	} else if len(args) == 1 {
		if bufSize, ok = args[0].(int); !ok {
			panic("bufSize must be a int")
		}
	} else {
		panic("Error parameters")
	}
	buff := make([]uint16, int(bufSize))
	ret, _, lastErr := controlGetHandleAsText.Call(strPtr(title), strPtr(text), strPtr(control), uintptr(unsafe.Pointer(&buff[0])), intPtr(bufSize))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return (goWString(buff))
}

//ControlGetPos -- Retrieves the position and size of a control relative to its window.
func ControlGetPos(title, text, control string) RECT {
	lprect := RECT{}
	ret, _, lastErr := controlGetPos.Call(strPtr(title), strPtr(text), strPtr(control), uintptr(unsafe.Pointer(&lprect)))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return lprect
}

//ControlGetPosByHandle -- Retrieves the position and size of a control relative to its window.
func ControlGetPosByHandle(title, text, control string) RECT {
	lprect := RECT{}
	ret, _, lastErr := controlGetPosByHandle.Call(strPtr(title), strPtr(text), strPtr(control), uintptr(unsafe.Pointer(&lprect)))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return lprect
}

//ControlGetText -- Retrieves text from a control.
func ControlGetText(title, text, control string, args ...interface{}) string {
	var bufSize int
	var ok bool

	if len(args) == 0 {
		bufSize = 256
	} else if len(args) == 1 {
		if bufSize, ok = args[0].(int); !ok {
			panic("bufSize must be a int")
		}
	} else {
		panic("Error parameters")
	}
	buff := make([]uint16, int(bufSize))
	ret, _, lastErr := controlGetText.Call(strPtr(title), strPtr(text), strPtr(control), uintptr(unsafe.Pointer(&buff[0])), intPtr(bufSize))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return (goWString(buff))
}

//ControlGetTextByHandle -- Retrieves text from a control.
func ControlGetTextByHandle(handle, control HWND, args ...interface{}) string {
	var bufSize int
	var ok bool

	if len(args) == 0 {
		bufSize = 256
	} else if len(args) == 1 {
		if bufSize, ok = args[0].(int); !ok {
			panic("bufSize must be a int")
		}
	} else {
		panic("Error parameters")
	}
	buff := make([]uint16, int(bufSize))
	ret, _, lastErr := controlGetTextByHandle.Call(uintptr(handle), uintptr(control), uintptr(unsafe.Pointer(&buff[0])), intPtr(bufSize))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return (goWString(buff))
}

//ControlHide -- Hides a control.
func ControlHide(title, text, control string) int {
	ret, _, lastErr := controlHide.Call(strPtr(title), strPtr(text), strPtr(control))
	if int(ret) == 1 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlHideByHandle -- Hides a control.
func ControlHideByHandle(title, text, control string) int {
	ret, _, lastErr := controlHideByHandle.Call(strPtr(title), strPtr(text), strPtr(control))
	if int(ret) == 1 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlMove -- Hides a control.
func ControlMove(title, text, control string, x, y int, args ...interface{}) int {
	var width, height int
	var ok bool

	if len(args) == 0 {
		width = -1
		height = -1
	} else if len(args) == 2 {
		if width, ok = args[0].(int); !ok {
			panic("width must be a int")
		}
		if height, ok = args[1].(int); !ok {
			panic("height must be a int")
		}
	} else {
		panic("Error parameters")
	}
	ret, _, lastErr := controlMove.Call(strPtr(title), strPtr(text), strPtr(control), intPtr(x), intPtr(y), intPtr(width), intPtr(height))
	if int(ret) == 1 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlMoveByHandle -- Hides a control.
func ControlMoveByHandle(handle, control HWND, x, y int, args ...interface{}) int {
	var width, height int
	var ok bool

	if len(args) == 0 {
		width = -1
		height = -1
	} else if len(args) == 2 {
		if width, ok = args[0].(int); !ok {
			panic("width must be a int")
		}
		if height, ok = args[1].(int); !ok {
			panic("height must be a int")
		}
	} else {
		panic("Error parameters")
	}
	ret, _, lastErr := controlMoveByHandle.Call(uintptr(handle), uintptr(control), intPtr(x), intPtr(y), intPtr(width), intPtr(height))
	if int(ret) == 1 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlShow -- Shows a control that was hidden.
func ControlShow(title, text, control string) int {
	ret, _, lastErr := controlShow.Call(strPtr(title), strPtr(text), strPtr(control))
	if int(ret) == 1 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlShowByHandle -- Shows a control that was hidden.
func ControlShowByHandle(handle, control HWND) int {
	ret, _, lastErr := controlShowByHandle.Call(uintptr(handle), uintptr(control))
	if int(ret) == 1 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ControlTreeView -- Sends a command to a TreeView32 control.
func ControlTreeView(title, text, control, command string, args ...interface{}) string {
	var Extra1, Extra2 string
	var bufSize int
	var ok bool

	if len(args) == 0 {
		Extra1 = ""
		Extra2 = ""
		bufSize = 256
	} else if len(args) == 1 {
		if Extra1, ok = args[0].(string); !ok {
			panic("Extra1 must be a string")
		}
		Extra2 = ""
		bufSize = 256
	} else if len(args) == 2 {
		if Extra1, ok = args[0].(string); !ok {
			panic("Extra1 must be a string")
		}
		if Extra2, ok = args[1].(string); !ok {
			panic("Extra2 must be a string")
		}
		bufSize = 256
	} else if len(args) == 3 {
		if Extra1, ok = args[0].(string); !ok {
			panic("Extra1 must be a string")
		}
		if Extra2, ok = args[1].(string); !ok {
			panic("Extra2 must be a string")
		}
		if bufSize, ok = args[2].(int); !ok {
			panic("bufSize must be a int")
		}
	} else {
		panic("Error parameters")
	}
	buff := make([]uint16, int(bufSize))
	ret, _, lastErr := controlTreeView.Call(strPtr(title), strPtr(text), strPtr(control), strPtr(command), strPtr(Extra1), strPtr(Extra2), uintptr(unsafe.Pointer(&buff[0])), intPtr(bufSize))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return (goWString(buff))
}

//ControlTreeViewByHandle -- Sends a command to a TreeView32 control.
func ControlTreeViewByHandle(handle, control HWND, command string, args ...interface{}) string {
	var Extra1, Extra2 string
	var bufSize int
	var ok bool

	if len(args) == 0 {
		Extra1 = ""
		Extra2 = ""
		bufSize = 256
	} else if len(args) == 1 {
		if Extra1, ok = args[0].(string); !ok {
			panic("Extra1 must be a string")
		}
		Extra2 = ""
		bufSize = 256
	} else if len(args) == 2 {
		if Extra1, ok = args[0].(string); !ok {
			panic("Extra1 must be a string")
		}
		if Extra2, ok = args[1].(string); !ok {
			panic("Extra2 must be a string")
		}
		bufSize = 256
	} else if len(args) == 3 {
		if Extra1, ok = args[0].(string); !ok {
			panic("Extra1 must be a string")
		}
		if Extra2, ok = args[1].(string); !ok {
			panic("Extra2 must be a string")
		}
		if bufSize, ok = args[2].(int); !ok {
			panic("bufSize must be a int")
		}
	} else {
		panic("Error parameters")
	}
	buff := make([]uint16, int(bufSize))
	ret, _, lastErr := controlTreeView.Call(uintptr(handle), uintptr(control), strPtr(command), strPtr(Extra1), strPtr(Extra2), uintptr(unsafe.Pointer(&buff[0])), intPtr(bufSize))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return (goWString(buff))
}

//Opt -- set option
func Opt(option, value string) int {
	ret, _, lastErr := opt.Call(strPtr(option), strPtr(value))
	if int(ret) == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

func findTermChr(buff []uint16) int {
	for i, char := range buff {
		if char == 0x0 {
			return i
		}
	}
	panic("not supposed to happen")
}

func intPtr(n int) uintptr {
	return uintptr(n)
}

func strPtr(s string) uintptr {
	return uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(s)))
}

// GoWString -- Convert a uint16 arrry C string to a Go String
func goWString(s []uint16) string {
	pos := findTermChr(s)
	// log.Println(string(utf16.Decode(s[0:pos])))
	return (string(utf16.Decode(s[0:pos])))
}
