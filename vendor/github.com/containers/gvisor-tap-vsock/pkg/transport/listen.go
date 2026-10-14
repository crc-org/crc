package transport

import (
	"errors"
	"net"
	"net/url"
	"os"
	"runtime"
)

// UnixSocketPath extracts the filesystem path from a parsed unix:// URL,
// handling the leading "/" that url.Parse adds before Windows drive letters
// (e.g. unix:///c:/path → Path="/c:/path" -> "c:/path").
// The goos parameter allows callers to specify the target OS for testability;
// pass runtime.GOOS for production use.
func UnixSocketPath(u *url.URL, goos string) string {
	path := u.Path
	if goos == "windows" && len(path) > 2 && path[0] == '/' && path[2] == ':' {
		path = path[1:]
	}
	return path
}

func defaultListenURL(url *url.URL) (net.Listener, error) {
	switch url.Scheme {
	case "unix":
		path := UnixSocketPath(url, runtime.GOOS)
		listener, err := net.Listen(url.Scheme, path)
		if err != nil {
			return nil, err
		}
		if err := os.Chmod(path, 0600); err != nil { // #nosec G703 - socket path from configured listen URL
			_ = listener.Close()
			return nil, err
		}
		return listener, nil
	case "tcp":
		return net.Listen("tcp", url.Host)
	default:
		return nil, errors.New("unexpected scheme")
	}
}

func Listen(endpoint string) (net.Listener, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return listenURL(parsed)
}
