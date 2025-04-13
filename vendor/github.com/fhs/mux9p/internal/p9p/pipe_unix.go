// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package p9p

import (
	"errors"
	"net"
	"syscall"
)

func pipe(p []int) error {
	if len(p) != 2 {
		return errors.New("bad argument to pipe")
	}
	fd, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	copy(p, fd[:])
	return nil
}

func sendFD(conn *net.UnixConn, fd uintptr) error {
	f, err := conn.File()
	if err != nil {
		return err
	}
	return syscall.Sendmsg(int(f.Fd()), nil, syscall.UnixRights(int(fd)), nil, 0)
}

func receiveFD(conn *net.UnixConn) (fd uintptr, err error) {
	f, err := conn.File()
	if err != nil {
		return 0, err
	}
	b := make([]byte, syscall.CmsgSpace(4))
	if _, _, _, _, err = syscall.Recvmsg(int(f.Fd()), nil, b, 0); err != nil {
		return 0, err
	}
	msgs, err := syscall.ParseSocketControlMessage(b)
	if err != nil {
		return 0, err
	}
	if len(msgs) == 0 {
		return 0, errors.New("no control message")
	}
	fds, err := syscall.ParseUnixRights(&msgs[0])
	if err != nil {
		return 0, err
	}
	if len(fds) == 0 {
		return 0, errors.New("no file descriptor")
	}
	return uintptr(fds[0]), nil
}
