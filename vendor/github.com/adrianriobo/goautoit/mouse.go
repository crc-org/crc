// +build windows
// +build amd64

package goautoit

import (
	"log"
	"unsafe"
)

//MouseClickDrag -- Perform a mouse click and drag operation.
func MouseClickDrag(button string, x1, y1, x2, y2 int, args ...interface{}) int {
	var nSpeed int
	var ok bool

	if len(args) == 0 {
		nSpeed = 10
	} else if len(args) == 1 {
		if nSpeed, ok = args[0].(int); !ok {
			panic("nSpeed must be a int")
		}
	} else {
		panic("Error parameters")
	}
	ret, _, lastErr := mouseClickDrag.Call(strPtr(button), intPtr(x1), intPtr(x2), intPtr(y2), intPtr(nSpeed))
	if int(ret) != 1 {
		log.Print("failure!!!")
		log.Println(lastErr)
	}
	return int(ret)
}

// MouseDown -- Perform a mouse down event at the current mouse position.
func MouseDown(args ...interface{}) int {
	var button string
	var ok bool

	if len(args) == 0 {
		button = DefaultMouseButton
	} else if len(args) == 1 {
		if button, ok = args[0].(string); !ok {
			panic("nSpeed must be a int")
		}
	} else {
		panic("Error parameters")
	}
	ret, _, lastErr := mouseDown.Call(strPtr(button))
	if int(ret) != 1 {
		log.Print("failure!!!")
		log.Println(lastErr)
	}
	return int(ret)
}

// MouseUp -- Perform a mouse up event at the current mouse position.
func MouseUp(args ...interface{}) int {
	var button string
	var ok bool

	if len(args) == 0 {
		button = DefaultMouseButton
	} else if len(args) == 1 {
		if button, ok = args[0].(string); !ok {
			panic("nSpeed must be a int")
		}
	} else {
		panic("Error parameters")
	}
	ret, _, lastErr := mouseUp.Call(strPtr(button))
	if int(ret) != 1 {
		log.Print("failure!!!")
		log.Println(lastErr)
	}
	return int(ret)
}

//MouseGetCursor -- Returns the cursor ID Number for the current Mouse Cursor.
func MouseGetCursor() int {
	ret, _, lastErr := mouseGetCursor.Call()
	if int(ret) == -1 {
		log.Println(lastErr)
	}
	return int(ret)
}

// MouseGetPos -- Retrieves the current position of the mouse cursor.
func MouseGetPos() (int32, int32) {
	var point = POINT{}
	ret, _, lastErr := mouseGetPos.Call(uintptr(unsafe.Pointer(&point)))
	if ret == 0 {
		log.Println(lastErr)
	}
	return point.X, point.Y
}

//MouseMove -- Moves the mouse pointer.
func MouseMove(x, y int, args ...interface{}) {
	var nSpeed int
	var ok bool

	if len(args) == 0 {
		nSpeed = -1
	} else if len(args) == 1 {
		if nSpeed, ok = args[0].(int); !ok {
			panic("nSpeed must be a int")
		}
	} else {
		panic("Error parameters")
	}
	ret, _, lastErr := mouseMove.Call(intPtr(x), intPtr(y), intPtr(nSpeed))
	if ret == 0 {
		log.Println(lastErr)
	}
}

//MouseWheel -- Moves the mouse wheel up or down.
func MouseWheel(szDirection string, args ...interface{}) int {
	var nClicks int
	var ok bool

	if len(args) == 0 {
		nClicks = 1
	} else if len(args) == 1 {
		if nClicks, ok = args[0].(int); !ok {
			panic("nClicks must be a int")
		}
	} else {
		panic("Error parameters")
	}
	ret, _, lastErr := mouseWheel.Call(strPtr(szDirection), intPtr(nClicks))
	if ret == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}
