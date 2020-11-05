package transport

import (
	"net"
	"net/url"
	"strconv"

	mdlayhervsock "github.com/mdlayher/vsock"
	"github.com/pkg/errors"
)

func Dial(endpoint string) (net.Conn, string, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, "", err
	}
	switch parsed.Scheme {
	case "vsock":
		contextID, err := strconv.Atoi(parsed.Hostname())
		if err != nil {
			return nil, "", err
		}
		port, err := strconv.Atoi(parsed.Port())
		if err != nil {
			return nil, "", err
		}
		conn, err := mdlayhervsock.Dial(uint32(contextID), uint32(port))
		return conn, parsed.Path, err
	case "unix":
		conn, err := net.Dial("unix", parsed.Path)
		return conn, "/connect", err
	default:
		return nil, "", errors.New("unexpected scheme")
	}
}
