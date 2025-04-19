//go:build linux || darwin
// +build linux darwin

package p9

import (
	"os"
	"os/user"
	"path/filepath"
)

// NamespaceDir returns the path of the directory that is used for the
// current namespace. On Unix-like systems, this is
// /tmp/ns.$USER.$DISPLAY.
//
// If looking up the current user's name fails, this function will
// panic.
func NamespaceDir() string {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	display, ok := os.LookupEnv("DISPLAY")
	if !ok {
		display = ":0"
	}

	return filepath.Join("/", "tmp", "ns."+u.Username+"."+display)
}
