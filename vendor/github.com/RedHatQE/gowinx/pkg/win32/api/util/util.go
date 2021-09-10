// +build windows
package util

import "syscall"

func MakeLPARAM(hiword uint16, loword uint16) uintptr {
	return uintptr((hiword << 16) | uint16(loword&0xffff))
}

func EvalSyscallHandler(r0 uintptr, e1 syscall.Errno) (hwnd syscall.Handle, err error) {
	hwnd = syscall.Handle(r0)
	if e1 != 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvalSyscallInt32(r0 uintptr, e1 syscall.Errno) (value int32, err error) {
	if r0 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	value = int32(r0)
	return
}

func EvalSyscallBool(r0 uintptr, e1 syscall.Errno) (value bool, err error) {
	if r0 == 0 {
		value = false
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	} else {
		value = true
	}
	return
}
