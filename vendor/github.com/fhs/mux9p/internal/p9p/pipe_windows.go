package p9p

import (
	"errors"
	"net"
)

var ErrUnsupported = errors.New("not supported on windows")

func pipe(p []int) error {
	return ErrUnsupported
}

func sendFD(conn *net.UnixConn, fd uintptr) error {
	return ErrUnsupported
}

func receiveFD(conn *net.UnixConn) (fd uintptr, err error) {
	return 0, ErrUnsupported
}
