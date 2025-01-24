//go:build !windows
// +build !windows

package shell

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/spf13/cast"
)

var (
	ErrUnknownShell        = errors.New("Error: Unknown shell")
	currentProcessSupplier = createCurrentProcess
)

// AbstractProcess is an interface created to abstract operations of the gopsutil library
// It is created so that we can override the behavior while writing unit tests by providing
// a mock implementation.
type AbstractProcess interface {
	Name() (string, error)
	Parent() (AbstractProcess, error)
}

// RealProcess is a wrapper implementation of AbstractProcess to wrap around the gopsutil library's
// process.Process object. This implementation is used in production code.
type RealProcess struct {
	*process.Process
}

func (p *RealProcess) Parent() (AbstractProcess, error) {
	parentProcess, err := p.Process.Parent()
	if err != nil {
		return nil, err
	}
	return &RealProcess{parentProcess}, nil
}

func createCurrentProcess() AbstractProcess {
	currentProcess, err := process.NewProcess(cast.ToInt32(os.Getpid()))
	if err != nil {
		return nil
	}
	return &RealProcess{currentProcess}
}

// detect detects user's current shell.
func detect() (string, error) {
	detectedShell := detectShellByCheckingProcessTree(currentProcessSupplier())
	if detectedShell == "" {
		fmt.Printf("The default lines below are for a sh/bash shell, you can specify the shell you're using, with the --shell flag.\n\n")
		return "", ErrUnknownShell
	}

	return filepath.Base(detectedShell), nil
}

// detectShellByCheckingProcessTree attempts to identify the shell being used by
// examining the process tree starting from the given process ID. This function
// traverses up to ProcessDepthLimit levels up the process hierarchy.
// Parameters:
//   - pid (int): The process ID to start checking from.
//
// Returns:
//   - string: The name of the shell if found (e.g., "zsh", "bash", "fish");
//     otherwise, an empty string is returned if no matching shell is detected
//     or an error occurs during the process tree traversal.
//
// Examples:
//
//	shellName := detectShellByCheckingProcessTree(1234)
func detectShellByCheckingProcessTree(p AbstractProcess) string {
	for p != nil {
		processName, err := p.Name()
		if err != nil {
			return ""
		}
		if processName == "zsh" || processName == "bash" || processName == "fish" {
			return processName
		}
		p, err = p.Parent()
		if err != nil {
			return ""
		}
	}
	return ""
}
