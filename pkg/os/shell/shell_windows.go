package shell

import (
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
)

var (
	supportedShell = []string{"cmd", "powershell", "wsl", "bash", "zsh", "fish"}
)

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
	case strings.Contains(strings.ToLower(shell), "bash"):
		return "bash"
	default:
		return defaultShell
	}
}

func detect() (string, error) {
	shell := detectShellByCheckingProcessTree(currentProcessSupplier())

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
		if len(lineParts) == 2 && slices.ContainsFunc(supportedShell, func(listElem string) bool {
			return strings.HasPrefix(lineParts[1], listElem)
		}) {
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
