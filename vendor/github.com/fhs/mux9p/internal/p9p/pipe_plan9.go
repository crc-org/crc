package p9p

import (
	"net"
	"syscall"
)

func pipe(p []int) error {
	return syscall.Pipe(p)
}

func sendFD(conn *net.UnixConn, fd uintptr) error {
	return syscall.EPLAN9
}

func receiveFD(conn *net.UnixConn) (fd uintptr, err error) {
	return 0, syscall.EPLAN9
}
