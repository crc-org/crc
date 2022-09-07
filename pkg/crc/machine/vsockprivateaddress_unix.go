//go:build !windows
// +build !windows

package machine

func VsockPrivateAddress() (addrStr string) {
	return "127.0.0.1"
}
