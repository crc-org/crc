// +build !plan9

package real

import (
	"fmt"
	"os"
	"os/user"
	"syscall"
)

func getUserGroup(info os.FileInfo) (string, string, error) {
	sysi := info.Sys()
	if sysi == nil {
		return "", "", fmt.Errorf("Cannot get system-specific info.")
	}
	sys := sysi.(*syscall.Stat_t)
	u, err := user.LookupId(fmt.Sprintf("%d", sys.Uid))
	if err != nil {
		return "", "", fmt.Errorf("Failed to lookup user: %s\n", err)
	}
	g, err := user.LookupGroupId(fmt.Sprintf("%d", sys.Gid))
	if err != nil {
		return "", "", fmt.Errorf("Failed to lookup group: %s\n", err)
	}
	return u.Username, g.Name, nil
}
