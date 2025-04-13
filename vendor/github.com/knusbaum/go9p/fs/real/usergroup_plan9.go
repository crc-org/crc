package real

import (
	"fmt"
	"os"
	"syscall"
)

func getUserGroup(info os.FileInfo) (string, string, error) {
	sysi := info.Sys()
	if sysi == nil {
		return "", "", fmt.Errorf("Cannot get system-specific info.")
	}
	sys := sysi.(*syscall.Dir)
	return sys.Uid, sys.Gid, nil
}
