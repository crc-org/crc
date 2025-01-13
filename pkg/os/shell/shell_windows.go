package shell

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
)

var (
	supportedShell = []string{"cmd", "powershell", "bash", "zsh", "fish"}
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
	case strings.Contains(strings.ToLower(shell), "wsl"):
		return detectShellByInvokingCommand("bash", "wsl", []string{"-e", "bash", "-c", "ps -ao pid=,comm="})
	case filepath.IsAbs(shell) && strings.Contains(strings.ToLower(shell), "bash"):
		return "bash"
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

// detectShellByInvokingCommand is a utility method that tries to detect current shell in use by invoking `ps` command.
// This method is extracted so that it could be used on Windows Subsystem for Linux (WSL). It executes
// the command provided in the method arguments and then passes the output to inspectProcessOutputForRecentlyUsedShell
// for evaluation.
//
// It receives two arguments:
// - defaultShell : default shell to revert back to in case it's unable to detect.
// - command: command to be executed
// - args: a string array containing command arguments
//
// It returns a string value representing current shell.
func detectShellByInvokingCommand(defaultShell string, command string, args []string) string {
	stdOut, _, err := CommandRunner.Run(command, args...)
	if err != nil {
		return defaultShell
	}

	detectedShell := inspectProcessOutputForRecentlyUsedShell(stdOut)
	if detectedShell == "" {
		return defaultShell
	}
	logging.Debugf("Detected shell: %s", detectedShell)
	return detectedShell
}

// inspectProcessOutputForRecentlyUsedShell inspects output of ps command to detect currently active shell session.
//
// It parses the output into a struct, filters process types by name then reverse sort it with pid and returns the first element.
//
// It takes one argument:
//
// - psCommandOutput: output of ps command executed on a particular shell session
//
// It returns:
//
//   - a string value (one of `zsh`, `bash` or `fish`) for current shell environment in use. If it's not able to determine
//     underlying shell type, it returns and empty string.
//
// This method tries to check all processes open and filters out shell sessions (one of `zsh`, `bash` or `fish)
// It then returns first shell process.
//
// For example, if ps command gives this output:
//
//	2908 ps
//	2889 fish
//	823 bash
//
// Then this method would return `fish` as it's the first shell process.
func inspectProcessOutputForRecentlyUsedShell(psCommandOutput string) string {
	type ProcessOutput struct {
		processID int
		output    string
	}
	var processOutputs []ProcessOutput
	lines := strings.Split(psCommandOutput, "\n")
	for _, line := range lines {
		lineParts := strings.Split(strings.TrimSpace(line), " ")
		if len(lineParts) == 2 && (strings.Contains(lineParts[1], "zsh") ||
			strings.Contains(lineParts[1], "bash") ||
			strings.Contains(lineParts[1], "fish")) {
			parsedProcessID, err := strconv.Atoi(lineParts[0])
			if err == nil {
				processOutputs = append(processOutputs, ProcessOutput{
					processID: parsedProcessID,
					output:    lineParts[1],
				})
			}
		}
	}
	// Reverse sort the processes by PID (higher to lower)
	sort.Slice(processOutputs, func(i, j int) bool {
		return processOutputs[i].processID > processOutputs[j].processID
	})

	if len(processOutputs) > 0 {
		return processOutputs[0].output
	}
	return ""
}
