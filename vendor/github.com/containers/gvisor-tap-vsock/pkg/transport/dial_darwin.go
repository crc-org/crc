package transport

import (
	"errors"
	"net"
)

func Dial(_ string) (net.Conn, string, error) {
	return nil, "", errors.New("unsupported")
}
