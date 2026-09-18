//go:build !linux

package machine

import (
	"fmt"
	"runtime"
)

func (client *client) GenerateBundle(_ bool) error {
	return fmt.Errorf("Not implemented for %s", runtime.GOOS)
}
