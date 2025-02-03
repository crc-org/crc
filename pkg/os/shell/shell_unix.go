//go:build !windows
// +build !windows

package shell

import (
	"fmt"
	"path/filepath"
)

// detect detects user's current shell.
func detect() (string, error) {
	detectedShell := detectShellByCheckingProcessTree(currentProcessSupplier())
	if detectedShell == "" {
		fmt.Printf("The default lines below are for a sh/bash shell, you can specify the shell you're using, with the --shell flag.\n\n")
		return "", ErrUnknownShell
	}

	return filepath.Base(detectedShell), nil
}
