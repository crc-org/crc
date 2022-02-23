//go:build !windows
// +build !windows

package sshclient

import (
	"errors"
	"net"
	"net/url"
)

func ListenNpipe(socketURI *url.URL) (net.Listener, error) {
	return nil, errors.New("named pipes are not supported by this platform")
}
