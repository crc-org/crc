package shell

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/spf13/cast"

	crcos "github.com/crc-org/crc/v2/pkg/os"
)

var (
	CommandRunner                           = crcos.NewLocalCommandRunner()
	WindowsSubsystemLinuxKernelMetadataFile = "/proc/version"
	ErrUnknownShell                         = errors.New("error: Unknown shell")
	currentProcessSupplier                  = createCurrentProcess
)

type Config struct {
	Prefix     string
	Delimiter  string
	Suffix     string
	PathSuffix string
}

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
		if slices.ContainsFunc(supportedShell, func(listElem string) bool {
			return strings.HasPrefix(processName, listElem)
		}) {
			return processName
		}
		p, err = p.Parent()
		if err != nil {
			return ""
		}
	}
	return ""
}
