package transport

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"strconv"
)

const DefaultURL = "vsock://null:1024/vm_directory"

func listenURL(parsed *url.URL) (net.Listener, error) {
	switch parsed.Scheme {
	case "vsock":
		port, err := strconv.Atoi(parsed.Port())
		if err != nil {
			return nil, err
		}
		path := path.Join(parsed.Path, fmt.Sprintf("00000002.%08x", port))
		listener, err := net.ListenUnix("unix", &net.UnixAddr{
			Name: path,
			Net:  "unix",
		})
		if err != nil {
			return nil, err
		}
		if err := os.Chmod(path, 0600); err != nil { // #nosec G703 - constructed path for socket permissions
			_ = listener.Close()
			return nil, err
		}
		return listener, nil
	default:
		return defaultListenURL(parsed)
	}
}
