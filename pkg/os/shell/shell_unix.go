//go:build !windows
// +build !windows

package shell

import (
	"errors"
	"fmt"
	"path/filepath"
)

var (
	ErrUnknownShell = errors.New("Error: Unknown shell")
)

// detect detects user's current shell.
func detect() (string, error) {
	detectedShell := detectShellByInvokingCommand("", "ps", []string{"-o", "pid=,comm="})
	if detectedShell == "" {
		fmt.Printf("The default lines below are for a sh/bash shell, you can specify the shell you're using, with the --shell flag.\n\n")
		return "", ErrUnknownShell
	}

	return filepath.Base(detectedShell), nil
}
