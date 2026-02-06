//go:build darwin

package transport

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"syscall"
)

type connectedUnixgramConn struct {
	*net.UnixConn
	remoteAddr *net.UnixAddr
}

func connectListeningUnixgramConn(conn *net.UnixConn, remoteAddr *net.UnixAddr) (*connectedUnixgramConn, error) {
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return nil, err
	}
	err = rawConn.Control(func(fd uintptr) {
		if err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF, 1*1024*1024); err != nil {
			return
		}
		if err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF, 4*1024*1024); err != nil {
			return
		}
	})
	if err != nil {
		return nil, err
	}

	return &connectedUnixgramConn{
		UnixConn:   conn,
		remoteAddr: remoteAddr,
	}, nil
}

func (conn *connectedUnixgramConn) RemoteAddr() net.Addr {
	return conn.remoteAddr
}

func (conn *connectedUnixgramConn) Write(b []byte) (int, error) {
	return conn.WriteTo(b, conn.remoteAddr)
}

func peekAddress(listeningConn *net.UnixConn) (*net.UnixAddr, error) {
	rawConn, err := listeningConn.SyscallConn()
	if err != nil {
		return nil, err
	}

	var vfkitSockaddr syscall.Sockaddr
	var getRemoteAddrErr error

	magic := make([]byte, 4)
	getRemoteAddr := func(fd uintptr) bool {
		_, vfkitSockaddr, getRemoteAddrErr = syscall.Recvfrom(int(fd), magic, syscall.MSG_PEEK|syscall.MSG_TRUNC)

		return !errors.Is(getRemoteAddrErr, syscall.EAGAIN)
	}
	if err := rawConn.Read(getRemoteAddr); err != nil {
		return nil, err
	}
	if getRemoteAddrErr != nil {
		return nil, getRemoteAddrErr
	}

	// If it's the old vfkit handshake, consume it
	if bytes.Equal(magic, []byte("VFKT")) {
		_, _, err = listeningConn.ReadFrom(magic)
		if err != nil {
			return nil, err
		}
	}

	vfkitSockaddrUnix, ok := vfkitSockaddr.(*syscall.SockaddrUnix)
	if !ok {
		return nil, fmt.Errorf("unexpected remote address type: %t", vfkitSockaddr)
	}
	if vfkitSockaddrUnix.Name == "" {
		return nil, fmt.Errorf("vfkit socket address is empty")
	}

	vfkitAddr := &net.UnixAddr{Name: vfkitSockaddrUnix.Name, Net: "unixgram"}
	return vfkitAddr, nil
}

func AcceptVfkit(listeningConn *net.UnixConn) (net.Conn, error) {
	peekedAddr, err := peekAddress(listeningConn)
	if err != nil {
		return nil, err
	}

	return connectListeningUnixgramConn(listeningConn, peekedAddr)
}
