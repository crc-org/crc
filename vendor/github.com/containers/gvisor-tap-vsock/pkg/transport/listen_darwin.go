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
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		return net.ListenUnix("unix", &net.UnixAddr{
			Name: path,
			Net:  "unix",
		})
	default:
		return defaultListenURL(parsed)
	}
}
