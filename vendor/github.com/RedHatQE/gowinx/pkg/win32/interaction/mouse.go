// +build windows
package interaction

import (
	"time"
	"unsafe"

	"github.com/RedHatQE/gowinx/pkg/util/logging"
	win32wam "github.com/RedHatQE/gowinx/pkg/win32/api/user-interface/windows-and-messages"
)

const (
	elementClickDelay time.Duration = 300 * time.Millisecond
)

type MOUSE_INPUT struct {
	Type uint32
	Mi   MOUSEINPUT
}

type MOUSEINPUT struct {
	Dx          int32
	Dy          int32
	MouseData   uint32
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

func ClickOnRect(rect win32wam.RECT) error {
	x := ((rect.Right - rect.Left) / 2) + rect.Left
	y := ((rect.Bottom - rect.Top) / 2) + rect.Top
	return Click(x, y)
}

func Click(x, y int32) error {
	events := [3]uint32{
		win32wam.MOUSEEVENTF_ABSOLUTE | win32wam.MOUSEEVENTF_MOVE,
		win32wam.MOUSEEVENTF_LEFTDOWN,
		win32wam.MOUSEEVENTF_LEFTUP}
	for _, event := range events {
		time.Sleep(elementClickDelay)
		if err := mouseInput(x, y, uint32(event)); err != nil {
			return err
		}
	}
	return nil
}

func mouseInput(x, y int32, dwFlags uint32) error {
	dx := getDX(x)
	dy := getDY(y)
	mouseInput := MOUSE_INPUT{
		Type: win32wam.INPUT_MOUSE,
		Mi: MOUSEINPUT{
			Dx:          dx,
			Dy:          dy,
			MouseData:   uint32(0),
			DwFlags:     dwFlags,
			Time:        uint32(0),
			DwExtraInfo: uintptr(0)}}

	actions := [1]MOUSE_INPUT{mouseInput}

	if success, err := win32wam.SendInput(uint32(2), unsafe.Pointer(&actions), int32(unsafe.Sizeof(mouseInput))); success > 0 {
		return nil
	} else {
		return err
	}
}

func getDX(x int32) int32 {
	return getDAxisValue(x, win32wam.SM_CXSCREEN)
}

func getDY(y int32) int32 {
	return getDAxisValue(y, win32wam.SM_CYSCREEN)
}

func getDAxisValue(axisValue, systemMetricConstant int32) int32 {
	if metric, err := win32wam.GetSystemMetrics(systemMetricConstant); err == nil {
		return (axisValue * 65536) / metric
	} else {
		logging.Errorf("Error getting D Axis value: %v", err)
		return 0
	}
}
