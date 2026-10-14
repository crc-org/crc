package transport

import (
	"net"
	"net/url"
	"os"
	"strconv"

	mdlayhervsock "github.com/mdlayher/vsock"
)

const DefaultURL = "vsock://:1024"

func listenURL(parsed *url.URL) (net.Listener, error) {
	switch parsed.Scheme {
	case "vsock":
		port, err := strconv.ParseUint(parsed.Port(), 10, 32)
		if err != nil {
			return nil, err
		}

		if parsed.Hostname() != "" {
			cid, err := strconv.ParseUint(parsed.Hostname(), 10, 32)
			if err != nil {
				return nil, err
			}
			return mdlayhervsock.ListenContextID(uint32(cid), uint32(port), nil)
		}

		return mdlayhervsock.Listen(uint32(port), nil)
	case "unixpacket":
		listener, err := net.Listen(parsed.Scheme, parsed.Path)
		if err != nil {
			return nil, err
		}
		if err := os.Chmod(parsed.Path, 0600); err != nil { // #nosec G703 - socket path from configured listen URL
			_ = listener.Close()
			return nil, err
		}
		return listener, nil
	default:
		return defaultListenURL(parsed)
	}
}
