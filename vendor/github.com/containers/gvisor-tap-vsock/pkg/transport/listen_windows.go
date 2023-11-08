package transport

import (
	"errors"
	"net"
	"net/url"

	"github.com/linuxkit/virtsock/pkg/hvsock"
)

const DefaultURL = "vsock://00000400-FACB-11E6-BD58-64006A7986D3"

func Listen(endpoint string) (net.Listener, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	switch parsed.Scheme {
	case "vsock":
		svcid, err := hvsock.GUIDFromString(parsed.Hostname())
		if err != nil {
			return nil, err
		}
		return hvsock.Listen(hvsock.Addr{
			VMID:      hvsock.GUIDWildcard,
			ServiceID: svcid,
		})
	case "unix":
		return net.Listen(parsed.Scheme, parsed.Path)
	case "tcp":
		return net.Listen("tcp", parsed.Host)
	default:
		return nil, errors.New("unexpected scheme")
	}
}

func ListenUnixgram(endpoint string) (net.Conn, error) {
	return nil, errors.New("unsupported 'unixgram' scheme")
}

func AcceptVfkit(listeningConn net.Conn) (net.Conn, error) {
	return nil, errors.New("vfkit is unsupported on Windows")
}
