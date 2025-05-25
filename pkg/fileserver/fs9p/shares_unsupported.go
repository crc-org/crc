//go:build !windows

package fs9p

import (
	"fmt"
)

func StartHvsockShares(mounts map[string]string) error {
	if len(mounts) == 0 {
		return nil
	}

	return fmt.Errorf("this platform does not support sharing directories")
}
