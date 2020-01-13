package preflight

import (
	"bytes"
	"fmt"
	"io/ioutil"
	goos "os"
	"os/exec"
	"path/filepath"
	"strings"
	goTemplate "text/template"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	dl "github.com/code-ready/crc/pkg/download"
	"github.com/code-ready/crc/pkg/embed"
	"github.com/code-ready/crc/pkg/extract"
	"github.com/code-ready/crc/pkg/os"
	"github.com/pkg/errors"
)

const (
	trayPlistTemplate = `<?xml version='1.0' encoding='UTF-8'?>
	<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
	<plist version='1.0'>
		<dict>
			<key>Label</key>
			<string>crc.tray</string>
			<key>ProgramArguments</key>
			<array>
				<string>{{ .BinaryPath }}</string>
			</array>
			<key>StandardOutPath</key>
			<string>{{ .StdOutFilePath }}</string>
			<key>Disabled</key>
			<false/>
			<key>RunAtLoad</key>
			<true/>
		</dict>
	</plist>`

	daemonPlistTemplate = `<?xml version='1.0' encoding='UTF-8'?>
	<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
	<plist version='1.0'>
		<dict>
			<key>Label</key>
			<string>crc.daemon</string>
			<key>ProgramArguments</key>
			<array>
				<string>{{ .BinaryPath }}</string>
				<string>daemon</string>
				<string>--log-level</string>
				<string>debug</string>
			</array>
			<key>StandardOutPath</key>
			<string>{{ .StdOutFilePath }}</string>
			<key>KeepAlive</key>
			<true/>
			<key>Disabled</key>
			<false/>
		</dict>
	</plist>`

	daemonAgentLabel = "crc.daemon"
	trayAgentLabel   = "crc.tray"
)

var (
	launchAgentDir       = filepath.Join(constants.GetHomeDir(), "Library", "LaunchAgents")
	daemonPlistFilePath  = filepath.Join(launchAgentDir, "crc.daemon.plist")
	trayPlistFilePath    = filepath.Join(launchAgentDir, "crc.tray.plist")
	stdOutFilePathDaemon = filepath.Join(constants.CrcBaseDir, ".crcd-agent.log")
	stdOutFilePathTray   = filepath.Join(constants.CrcBaseDir, ".crct-agent.log")
)

type AgentConfig struct {
	BinaryPath     string
	StdOutFilePath string
}

func checkTrayExistsAndRunning() error {
	logging.Debug("Checking if daemon plist file exists")
	if !os.FileExists(daemonPlistFilePath) {
		return errors.New("Daemon plist file does not exist")
	}
	logging.Debug("Checking if crc agent running")
	if !agentRunning(daemonAgentLabel) {
		return errors.New("crc daemon is not running")
	}
	logging.Debug("Checking if tray plist file exists")
	if !os.FileExists(trayPlistFilePath) {
		return errors.New("Tray plist file does not exist")
	}
	logging.Debug("Checking if tray agent running")
	if !agentRunning(trayAgentLabel) {
		return errors.New("Tray is not running")
	}
	return nil
}

func fixTrayExistsAndRunning() error {
	// get the tray app
	if !trayAppCached() {
		err := downloadOrExtractTrayApp()
		if err != nil {
			return err
		}
	}
	currentExecutablePath, err := goos.Executable()
	if err != nil {
		return err
	}
	daemonConfig := AgentConfig{
		BinaryPath:     currentExecutablePath,
		StdOutFilePath: stdOutFilePathDaemon,
	}

	trayConfig := AgentConfig{
		BinaryPath:     filepath.Join(constants.CrcBinDir, constants.TrayBinaryName, "Contents", "MacOS", "CodeReady Containers"),
		StdOutFilePath: stdOutFilePathTray,
	}
	logging.Debug("Creating daemon plist")
	err = createPlist(daemonPlistTemplate, daemonConfig, daemonPlistFilePath)
	if err != nil {
		return err
	}
	logging.Debug("Creating tray plist")
	err = createPlist(trayPlistTemplate, trayConfig, trayPlistFilePath)
	if err != nil {
		return err
	}

	// load crc daemon
	err = launchctlLoadPlist(daemonPlistFilePath)
	if err != nil {
		return err
	}
	if !agentRunning(daemonAgentLabel) {
		if err = startAgent(daemonAgentLabel); err != nil {
			return err
		}
	}
	// load tray
	err = launchctlLoadPlist(trayPlistFilePath)
	if err != nil {
		return err
	}
	if !agentRunning(trayAgentLabel) {
		if err = startAgent(trayAgentLabel); err != nil {
			return err
		}
	}
	return nil
}

func createPlist(template string, config AgentConfig, plistPath string) error {
	var plistContent bytes.Buffer
	t, err := goTemplate.New("plist").Parse(template)
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

func launchctlLoadPlist(plistFilePath string) error {
	_, err := exec.Command("launchctl", "load", plistFilePath).Output() // #nosec G204
	return err
}

func startAgent(label string) error {
	_, err := exec.Command("launchctl", "start", label).Output() // #nosec G204
	return err
}

// check if a service (daemon,tray) is running
func agentRunning(label string) bool {
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

func trayAppCached() bool {
	return os.FileExists(constants.TrayBinaryPath)
}

func downloadOrExtractTrayApp() error {
	if trayAppCached() {
		return nil
	}

	// Extract the tray and put it in the bin directory.
	tmpArchivePath, err := ioutil.TempDir("", "crc")
	if err != nil {
		logging.Error("Failed creating temporary directory for extracting tray")
		return err
	}
	defer func() {
		_ = goos.RemoveAll(tmpArchivePath)
	}()

	logging.Debug("Trying to extract tray from crc binary")
	err = embed.Extract(filepath.Base(constants.GetCrcTrayDownloadURL()), tmpArchivePath)
	if err != nil {
		logging.Debug("Downloading crc tray")
		_, err = dl.Download(constants.GetCrcTrayDownloadURL(), tmpArchivePath, 0600)
		if err != nil {
			return err
		}
	}
	archivePath := filepath.Join(tmpArchivePath, filepath.Base(constants.GetCrcTrayDownloadURL()))
	outputPath := constants.CrcBinDir
	err = goos.MkdirAll(outputPath, 0750)
	if err != nil && !goos.IsExist(err) {
		return errors.Wrap(err, "Cannot create the target directory.")
	}
	err = extract.Uncompress(archivePath, outputPath)
	if err != nil {
		return errors.Wrapf(err, "Cannot uncompress '%s'", archivePath)
	}
	return nil
}
