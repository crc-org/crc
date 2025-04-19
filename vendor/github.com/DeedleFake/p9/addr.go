package p9

import (
	"path/filepath"
	"strings"
)

const standardPort = "564"

// ParseAddr returns a network and address pair for a basic address
// string. The parsing is done like this:
//
// If the address starts with "$", it is assumed to be a namespace and
// is located using GetNamespace.
//
// If the address starts with "./" or "/", it is assumed to be a path
// to a Unix socket.
//
// If the address contains a ":", it is a assumed to be a TCP address
// and port combo. As a special case, the pseudo-ports 9p and 9fs map
// to the standard 9P port number.
//
// If the address contains a single "!", it is assumed to be a network
// and address combo. If the network type is TCP, the standard 9P port
// is assumed.
//
// If the address contains two "!", it is assumed to be a network,
// address, and port, in that order.
//
// In all other cases, it is assumed to be only the address
// specificiation of a TCP address and the standard port is assumed.
func ParseAddr(addr string) (network, address string) {
	switch {
	case IsNamespaceAddr(addr):
		return GetNamespace(addr[1:])

	case strings.HasPrefix(addr, "./"), strings.HasPrefix(addr, "/"):
		return "unix", addr
	}

	parts := strings.SplitN(addr, ":", 2)
	if len(parts) == 2 {
		if (parts[1] == "9p") || (parts[1] == "9fs") {
			parts[1] = standardPort
		}

		return "tcp", strings.Join(parts, ":")
	}

	parts = strings.SplitN(addr, "!", 3)
	switch len(parts) {
	case 2:
		if parts[0] == "tcp" {
			parts[1] += ":" + standardPort
		}
		return parts[0], parts[1]

	case 3:
		if (parts[2] == "9p") || (parts[2] == "9fs") {
			parts[2] = standardPort
		}
		return parts[0], strings.Join(parts[1:], ":")
	}

	return "tcp", addr + ":" + standardPort
}

// GetNamespace returns the network and address that should be used to
// connect to the named program in the current namespace.
func GetNamespace(name string) (network, addr string) {
	return "unix", filepath.Join(NamespaceDir(), name)
}

// IsNamespaceAddr returns true if the given address would result in a
// namespace network/address pair if passed to ParseAddr.
func IsNamespaceAddr(addr string) bool {
	return strings.HasPrefix(addr, "$")
}
