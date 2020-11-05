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
	default:
		return nil, errors.New("unexpected scheme")
	}
}
