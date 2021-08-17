package transport

import (
	"net"

	"github.com/pkg/errors"
)

func Dial(endpoint string) (net.Conn, string, error) {
	return nil, "", errors.New("unsupported")
}
