package launchd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	goos "os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/os"
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
				<string>{{ .BinaryPath }}</string>
			{{ range .Args }}
				<string>{{ . }}</string>
			{{ end }}
			</array>
			<key>StandardOutPath</key>
			<string>{{ .StdOutFilePath }}</string>
			<key>Disabled</key>
			<false/>
			<key>RunAtLoad</key>
			<true/>
		</dict>
	</plist>`
)

// AgentConfig is struct to contain configuration for agent plist file
type AgentConfig struct {
	Label          string
	BinaryPath     string
	StdOutFilePath string
	Args           []string
}

var (
	launchAgentsDir = filepath.Join(constants.GetHomeDir(), "Library", "LaunchAgents")
)

func ensureLaunchAgentsDirExists() error {
	if err := goos.MkdirAll(launchAgentsDir, 0700); err != nil {
		return err
	}
	return nil
}

func getPlistPath(label string) string {
	plistName := fmt.Sprintf("%s.plist", label)

	return filepath.Join(launchAgentsDir, plistName)
}

// CreatePlist creates a launchd agent plist config file
func CreatePlist(config AgentConfig) error {
	if err := ensureLaunchAgentsDirExists(); err != nil {
		return err
	}

	var plistContent bytes.Buffer
	t, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return err
	}
	err = t.Execute(&plistContent, config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(getPlistPath(config.Label), plistContent.Bytes(), 0644)
	return err
}

func PlistExists(label string) bool {
	return os.FileExists(getPlistPath(label))
}

// LoadPlist loads a launchd agents' plist file
func LoadPlist(label string) error {
	return exec.Command("launchctl", "load", getPlistPath(label)).Run() // #nosec G204
}

// StartAgent starts a launchd agent
func StartAgent(label string) error {
	return exec.Command("launchctl", "start", label).Run() // #nosec G204
}

// StopAgent stops a launchd agent
func StopAgent(label string) error {
	return exec.Command("launchctl", "stop", label).Run() // #nosec G204
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
	out, err := exec.Command("bash", "-c", cmd).Output() // #nosec G204
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
	return exec.Command("launchctl", "remove", label).Run() // #nosec G204
}
