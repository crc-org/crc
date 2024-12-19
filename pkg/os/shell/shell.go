package shell

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	crcos "github.com/crc-org/crc/v2/pkg/os"
)

var (
	CommandRunner                           = crcos.NewLocalCommandRunner()
	WindowsSubsystemLinuxKernelMetadataFile = "/proc/version"
)

type Config struct {
	Prefix     string
	Delimiter  string
	Suffix     string
	PathSuffix string
}

func GetShell(userShell string) (string, error) {
	if userShell != "" {
		if !isSupportedShell(userShell) {
			return "", fmt.Errorf("'%s' is not a supported shell.\nSupported shells are %s", userShell, strings.Join(supportedShell, ", "))
		}
		return userShell, nil
	}
	return detect()
}

func isSupportedShell(userShell string) bool {
	for _, shell := range supportedShell {
		if userShell == shell {
			return true
		}
	}
	return false
}

func GenerateUsageHintWithComment(userShell, cmdLine string) string {
	return fmt.Sprintf("%s Run this command to configure your shell:\n%s %s",
		comment(userShell),
		comment(userShell),
		GenerateUsageHint(userShell, cmdLine))
}

func comment(userShell string) string {
	if userShell == "cmd" {
		return "REM"
	}
	return "#"
}

func GenerateUsageHint(userShell, cmdLine string) string {
	switch userShell {
	case "fish":
		return fmt.Sprintf("eval (%s)", cmdLine)
	case "powershell":
		return fmt.Sprintf("& %s | Invoke-Expression", cmdLine)
	case "cmd":
		return fmt.Sprintf("@FOR /f \"tokens=*\" %%i IN ('%s') DO @call %%i", cmdLine)
	default:
		return fmt.Sprintf("eval $(%s)", cmdLine)
	}
}

func GetEnvString(userShell string, envName string, envValue string) string {
	switch userShell {
	case "powershell":
		return fmt.Sprintf("$Env:%s = \"%s\"", envName, envValue)
	case "cmd":
		return fmt.Sprintf("SET %s=%s", envName, envValue)
	case "fish":
		return fmt.Sprintf("contains %s $fish_user_paths; or set -U fish_user_paths %s $fish_user_paths", convertToLinuxStylePath(userShell, envValue), convertToLinuxStylePath(userShell, envValue))
	default:
		return fmt.Sprintf("export %s=\"%s\"", envName, convertToLinuxStylePath(userShell, envValue))
	}
}

func GetPathEnvString(userShell string, prependedPath string) string {
	var pathStr string
	switch userShell {
	case "fish":
		pathStr = prependedPath
	case "powershell":
		pathStr = fmt.Sprintf("%s;$Env:PATH", prependedPath)
	case "cmd":
		pathStr = fmt.Sprintf("%s;%%PATH%%", prependedPath)
	default:
		pathStr = fmt.Sprintf("%s:$PATH", convertToLinuxStylePath(userShell, prependedPath))
	}

	return GetEnvString(userShell, "PATH", pathStr)
}

// convertToLinuxStylePath is a utility method to translate Windows paths to Linux environments (e.g. Git Bash).
//
// It receives two arguments:
// - userShell : currently active shell
// - path : Windows path to be converted
//
// It returns Linux equivalent of the Windows path.
//
// For example, a Windows path like `C:\Users\foo\.crc\bin\oc` is converted into `/C/Users/foo/.crc/bin/oc`.
func convertToLinuxStylePath(userShell string, path string) string {
	if IsWindowsSubsystemLinux() {
		return convertToWindowsSubsystemLinuxPath(path)
	}
	if strings.Contains(path, "\\") &&
		(userShell == "bash" || userShell == "zsh" || userShell == "fish") {
		path = strings.ReplaceAll(path, ":", "")
		path = strings.ReplaceAll(path, "\\", "/")

		return fmt.Sprintf("/%s", path)
	}
	return path
}

// convertToWindowsSubsystemLinuxPath is a utility method to translate between Windows and WSL(Windows Subsystem for
// Linux) paths. It relies on `wslpath` command to perform this conversion.
//
// It receives one argument:
// - path : Windows path to be converted to WSL path
//
// It returns translated WSL equivalent of provided windows path.
func convertToWindowsSubsystemLinuxPath(path string) string {
	stdOut, _, err := CommandRunner.Run("wsl", "-e", "bash", "-c", fmt.Sprintf("wslpath -a '%s'", path))
	if err != nil {
		return path
	}
	return strings.TrimSpace(stdOut)
}

// IsWindowsSubsystemLinux detects whether current system is using Windows Subsystem for Linux or not
//
// It checks for these conditions to make sure that current system has WSL installed:
// - `/proc/version` file is present
// - `/proc/version` file contents contain keywords `Microsoft` and `WSL`
//
// It above conditions are met, then this method returns `true` otherwise `false`.
func IsWindowsSubsystemLinux() bool {
	procVersionContent, err := os.ReadFile(WindowsSubsystemLinuxKernelMetadataFile)
	if err != nil {
		return false
	}
	if strings.Contains(string(procVersionContent), "Microsoft") ||
		strings.Contains(string(procVersionContent), "WSL") {
		return true
	}
	return false
}

// detectShellByInvokingCommand is a utility method that tries to detect current shell in use by invoking `ps` command.
// This method is extracted so that it could be used by unix systems as well as Windows (in case of WSL). It executes
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
	return detectedShell
}

// inspectProcessOutputForRecentlyUsedShell inspects output of ps command to detect currently active shell session.
//
// Note : This method assumes that ps command has already sorted the processes by `pid` in reverse order.
// It parses the output into a struct, filters process types by name and returns the first element.
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
	if len(processOutputs) > 0 {
		return processOutputs[0].output
	}
	return ""
}
