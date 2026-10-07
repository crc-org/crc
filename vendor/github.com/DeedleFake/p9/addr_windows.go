//go:build windows

package p9

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// NamespaceDir returns the path of the directory that is used for the
// current namespace. On Windows, domain name is included, if available.
//
// If looking up the current user's name fails, this function will
// panic.
func NamespaceDir() string {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	// On Windows, u.Username may contain the domain name as well (DOMAIN\user)
	domain, user, found := strings.Cut(u.Username, "\\")
	namespace := domain // will be the username if domain is not present
	if found {
		namespace = user + "." + domain
	}

	return filepath.Join(os.TempDir(), "ns."+namespace)
}
