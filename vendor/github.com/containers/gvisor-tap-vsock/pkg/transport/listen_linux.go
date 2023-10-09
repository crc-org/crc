package transport

import (
	"net"
	"net/url"
	"strconv"

	mdlayhervsock "github.com/mdlayher/vsock"
)

const DefaultURL = "vsock://:1024"

func listenURL(parsed *url.URL) (net.Listener, error) {
	switch parsed.Scheme {
	case "vsock":
		port, err := strconv.Atoi(parsed.Port())
		if err != nil {
			return nil, err
		}
		return mdlayhervsock.Listen(uint32(port), nil)
	case "unixpacket":
		return net.Listen(parsed.Scheme, parsed.Path)
	default:
		return defaultListenURL(parsed)
	}
}
