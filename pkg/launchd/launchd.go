package launchd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"text/template"
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

// CreatePlist creates a launchd agent plist config file
func CreatePlist(config AgentConfig, plistPath string) error {
	var plistContent bytes.Buffer
	t, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return err
	}
	err = t.Execute(&plistContent, config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(plistPath, plistContent.Bytes(), 0644)
	return err
}

// LoadPlist loads a launchd agents' plist file
func LoadPlist(plistFilePath string) error {
	return exec.Command("launchctl", "load", plistFilePath).Run() // #nosec G204
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
	// is running, otherwise returns "-"
	launchctlListCommand := `launchctl list | grep %s | awk '{print $1}'`
	cmd := fmt.Sprintf(launchctlListCommand, label)
	out, err := exec.Command("bash", "-c", cmd).Output() // #nosec G204
	if err != nil {
		return false
	}
	if strings.TrimSpace(string(out)) == "-" {
		return false
	}
	return true
}
