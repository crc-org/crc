package shell

import (
	"fmt"
	"math"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

var (
	supportedShell = []string{"cmd", "powershell"}
)

// re-implementation of private function in https://github.com/golang/go/blob/master/src/syscall/syscall_windows.go
func getProcessEntry(pid uint32) (pe *syscall.ProcessEntry32, err error) {
	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = syscall.CloseHandle(syscall.Handle(snapshot))
	}()

	var processEntry syscall.ProcessEntry32
	processEntry.Size = uint32(unsafe.Sizeof(processEntry))
	err = syscall.Process32First(snapshot, &processEntry)
	if err != nil {
		return nil, err
	}

	for {
		if processEntry.ProcessID == pid {
			pe = &processEntry
			return
		}

		err = syscall.Process32Next(snapshot, &processEntry)
		if err != nil {
			return nil, err
		}
	}
}

// getNameAndItsPpid returns the exe file name its parent process id.
func getNameAndItsPpid(pid uint32) (exefile string, parentid uint32, err error) {
	pe, err := getProcessEntry(pid)
	if err != nil {
		return "", 0, err
	}

	name := syscall.UTF16ToString(pe.ExeFile[:])
	return name, pe.ParentProcessID, nil
}

func shellType(shell string, defaultShell string) string {
	switch {
	case strings.Contains(strings.ToLower(shell), "powershell"):
		return "powershell"
	case strings.Contains(strings.ToLower(shell), "pwsh"):
		return "powershell"
	case strings.Contains(strings.ToLower(shell), "cmd"):
		return "cmd"
	default:
		return defaultShell
	}
}

func detect() (string, error) {
	shell := os.Getenv("SHELL")

	if shell == "" {
		pid := os.Getppid()
		if pid < 0 || pid > math.MaxUint32 {
			return "", fmt.Errorf("integer overflow for pid: %v", pid)
		}
		shell, shellppid, err := getNameAndItsPpid(uint32(pid))
		if err != nil {
			return "cmd", err // defaulting to cmd
		}
		shell = shellType(shell, "")
		if shell == "" {
			shell, _, err := getNameAndItsPpid(shellppid)
			if err != nil {
				return "cmd", err // defaulting to cmd
			}
			return shellType(shell, "cmd"), nil
		}
		return shell, nil
	}

	if os.Getenv("__fish_bin_dir") != "" {
		return "fish", nil
	}

	return shellType(shell, "cmd"), nil
}
