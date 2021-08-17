package transport

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"strconv"
)

const DefaultURL = "vsock://null:1024/vm_directory"

func Listen(endpoint string) (net.Listener, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
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
	case "unix":
		return net.Listen("unix", parsed.Path)
	case "tcp":
		return net.Listen("tcp", parsed.Host)
	default:
		return nil, errors.New("unexpected scheme")
	}
}
