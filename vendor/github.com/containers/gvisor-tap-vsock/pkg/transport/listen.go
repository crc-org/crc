package transport

import (
	"errors"
	"net"
	"net/url"
)

func defaultListenURL(url *url.URL) (net.Listener, error) {
	switch url.Scheme {
	case "unix":
		return net.Listen(url.Scheme, url.Path)
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
