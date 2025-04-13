// Package p9p implements some Plan 9 Port functions.
package p9p

import (
	"net"
	"os"
)

// Pipe returns a two-way pipe. Unlike os.Pipe, read and write is supported on both ends.
func Pipe() (*os.File, *os.File, error) {
	var p [2]int
	if err := pipe(p[:]); err != nil {
		return nil, nil, err
	}
	return os.NewFile(uintptr(p[0]), "|0"), os.NewFile(uintptr(p[1]), "|1"), nil
}

// SendFD sends a file descriptor through a unix domain socket.
func SendFD(conn *net.UnixConn, fd uintptr) error {
	return sendFD(conn, fd)
}

// ReceiveFD receives a file descriptor through a unix domain socket.
func ReceiveFD(conn *net.UnixConn) (fd uintptr, err error) {
	return receiveFD(conn)
}
