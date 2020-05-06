package service

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/os/windows/powershell"
)

func IsInstalled(serviceName string) bool {
	cmd := fmt.Sprintf("Get-Service -Name \"%s\"", serviceName)
	_, stdErr, err := powershell.Execute(cmd)
	if err != nil {
		logging.Debugf("Failed to execute powershell command: %v", err)
		return false
	}
	if strings.Contains(stdErr, serviceName) {
		return false
	}
	return true
}

func IsRunning(serviceName string) bool {
	cmd := fmt.Sprintf("(Get-Service -Name \"%s\").Status", serviceName)
	stdOut, _, err := powershell.Execute(cmd)
	if err != nil {
		logging.Debugf("Failed to execute powershell command: %v", err)
		return false
	}
	if strings.TrimSpace(stdOut) == "Running" {
		return true
	}
	return false
}

func Start(serviceName string) error {
	cmd := fmt.Sprintf("%s start \"%s\"", getScExePath(), serviceName)
	_, _, err := powershell.ExecuteAsAdmin(fmt.Sprintf("Starting service: %s", serviceName), cmd)
	if err != nil {
		logging.Debug("Failed to execute powershell command as admin")
		return err
	}
	// Needs a bit of time to reflect the service status in svc manager
	time.Sleep(2 * time.Second)

	// Since ExecuteAsAdmin doesn't return useful stdOut or stdErr
	// We independently check if the service was actually started
	if IsRunning(serviceName) {
		return nil
	}
	return fmt.Errorf("Could not start service: %s", serviceName)
}

func Stop(serviceName string) error {
	cmd := fmt.Sprintf("%s stop \"%s\"", getScExePath(), serviceName)
	_, _, err := powershell.ExecuteAsAdmin(fmt.Sprintf("Stopping service: %s", serviceName), cmd)
	if err != nil {
		logging.Debug("Failed to execute powershell command as admin")
		return err
	}
	time.Sleep(2 * time.Second)
	if IsRunning(serviceName) {
		return fmt.Errorf("Could not stop service: %s", serviceName)
	}
	return nil
}

func Delete(serviceName string) error {
	cmd := fmt.Sprintf("%s delete \"%s\"", getScExePath(), serviceName)
	_, _, err := powershell.ExecuteAsAdmin(fmt.Sprintf("Uninstalling service: %s", serviceName), cmd)
	if err != nil {
		logging.Debug("Failed to execute powershell command as admin")
		return err
	}
	time.Sleep(2 * time.Second)
	if IsInstalled(serviceName) {
		return fmt.Errorf("Could not uninstall service: %s", serviceName)
	}
	return nil
}

func Create(serviceName, binPath, accountName, password string) error {
	var args = []string{
		fmt.Sprintf("binPath= \"%s\"", binPath),
		"type= own",
		"start= auto",
		fmt.Sprintf("DisplayName= \"%s\"", serviceName),
		fmt.Sprintf("obj= \"%s\"", accountName),
		fmt.Sprintf("password= %s", password),
	}

	cmd := fmt.Sprintf("%s create --%% \"%s\" %s", getScExePath(), serviceName, strings.Join(args, " "))
	_, _, err := powershell.ExecuteAsAdmin("Install daemon service", cmd)
	if err != nil {
		logging.Debug("Failed to execute powershell command as admin")
		return fmt.Errorf("Error installing service %s: %v", serviceName, err)
	}
	time.Sleep(2 * time.Second)
	if !IsInstalled(serviceName) {
		return fmt.Errorf("Failed to install service: %s", serviceName)
	}
	return nil
}

func getScExePath() string {
	path, err := exec.LookPath("sc.exe")
	if err != nil {
		logging.Debug("Unable to find sc.exe on path")
		return ""
	}
	return path
}
