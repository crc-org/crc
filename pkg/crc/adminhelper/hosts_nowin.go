// +build !windows

package adminhelper

import (
	crcos "github.com/code-ready/crc/pkg/os"
)

func execute(args ...string) error {
	_, _, err := crcos.RunWithDefaultLocale(goodhostPath, args...)
	return err
}
