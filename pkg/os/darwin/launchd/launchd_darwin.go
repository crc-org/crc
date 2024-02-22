package launchd

import (
	"bytes"
	"errors"
	"fmt"
	goos "os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/os"
)

const (
	plistTemplate = `<?xml version='1.0' encoding='UTF-8'?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version='1.0'>
<dict>
	<key>Label</key>
	<string>{{ .Label }}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{ .ExecutablePath }}</string>
	{{ range .Args }}
		<string>{{ . }}</string>
	{{ end }}
	</array>
	<key>StandardOutPath</key>
	<string>{{ .StdOutFilePath }}</string>
	<key>StandardErrorPath</key>
	<string>{{ .StdErrFilePath }}</string>
	<key>EnvironmentVariables</key>
	<dict>
		{{ range $key, $value := .Env }}
		<key>{{ $key }}</key>
		<string>{{ $value }}</string>
		{{ end }}
	</dict>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>
`
)

// AgentConfig is struct to contain configuration for agent plist file
type AgentConfig struct {
	Label          string
	ExecutablePath string
	StdOutFilePath string
	StdErrFilePath string
	Args           []string
	Env            map[string]string
}

var (
	launchAgentsDir = filepath.Join(constants.GetHomeDir(), "Library", "LaunchAgents")
)

func ensureLaunchAgentsDirExists() error {
	return goos.MkdirAll(launchAgentsDir, 0700)
}

func getPlistPath(label string) string {
	plistName := fmt.Sprintf("%s.plist", label)

	return filepath.Join(launchAgentsDir, plistName)
}

func generatePlistContent(config AgentConfig) ([]byte, error) {
	var plistContent bytes.Buffer
	t, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return nil, err
	}
	err = t.Execute(&plistContent, config)
	if err != nil {
		return nil, err
	}

	return plistContent.Bytes(), nil
}

// CreatePlist creates a launchd agent plist config file
func CreatePlist(config AgentConfig) error {
	if err := ensureLaunchAgentsDirExists(); err != nil {
		return err
	}

	plist, err := generatePlistContent(config)
	if err != nil {
		return err
	}
	err = goos.WriteFile(getPlistPath(config.Label), plist, 0600)
	return err
}

func PlistExists(label string) bool {
	return os.FileExists(getPlistPath(label))
}

func CheckPlist(config AgentConfig) error {
	plist, err := generatePlistContent(config)
	if err != nil {
		return err
	}
	return os.FileContentMatches(getPlistPath(config.Label), plist)
}

// LoadPlist loads a launchd agents' plist file
func LoadPlist(label string) error {
	return runLaunchCtl("load", "-w", getPlistPath(label))
}

// UnloadPlist Unloads a launchd agent's service
func UnloadPlist(label string) error {
	return runLaunchCtl("unload", "-w", getPlistPath(label))
}

// RemovePlist removes a launchd agent plist config file
func RemovePlist(label string) error {
	if _, err := goos.Stat(getPlistPath(label)); !goos.IsNotExist(err) {
		return goos.Remove(getPlistPath(label))
	}
	return nil
}

// StartAgent starts a launchd agent
func StartAgent(label string) error {
	return runLaunchCtl("start", label)
}

// StopAgent stops a launchd agent
func StopAgent(label string) error {
	return runLaunchCtl("stop", label)
}

// RestartAgent restarts a launchd agent
func RestartAgent(label string) error {
	err := StopAgent(label)
	if err != nil {
		return err
	}
	return StartAgent(label)
}

// AgentRunning checks if a launchd service is running
func AgentRunning(label string) bool {
	// This command return a PID if the process
	// is running, otherwise returns "-" or empty
	// output if the agent is not loaded in launchd
	launchctlListCommand := `launchctl list | grep %s | awk '{print $1}'`
	cmd := fmt.Sprintf(launchctlListCommand, label)
	out, _, err := os.RunWithDefaultLocale("bash", "-c", cmd)
	if err != nil {
		return false
	}
	// match PID
	if match, err := regexp.MatchString(`^\d+$`, strings.TrimSpace(string(out))); err == nil && match {
		return true
	}
	return false
}

// Remove removes the agent from launchd
func Remove(label string) error {
	return runLaunchCtl("remove", label)
}

func runLaunchCtl(args ...string) error {
	_, _, err := os.RunWithDefaultLocale("launchctl", args...)
	return exitCodeToError(err)
}

func exitCodeToError(err error) error {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return err
	}

	stdout, _, localErr := os.RunWithDefaultLocale("launchctl", "error", fmt.Sprintf("%d", exitErr.ExitCode()))
	if localErr != nil {
		return err
	}

	return errors.New(stdout)
}
