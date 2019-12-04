package preflight

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"

	winnet "github.com/code-ready/crc/pkg/os/windows/network"
	"github.com/code-ready/crc/pkg/os/windows/powershell"

	"github.com/code-ready/crc/pkg/crc/machine/hyperv"
)

const (
	// Fall Creators update comes with the "Default Switch"
	minimumWindowsReleaseId = 1709
)

func checkVersionOfWindowsUpdate() error {
	windowsReleaseId := `(Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion" -Name ReleaseId).ReleaseId`

	stdOut, _, err := powershell.Execute(windowsReleaseId)
	if err != nil {
		logging.Debug(err.Error())
		return errors.New("Failed to get Windows release id")
	}
	yourWindowsReleaseId, err := strconv.Atoi(strings.TrimSpace(stdOut))

	if err != nil {
		logging.Debug(err.Error())
		return errors.Newf("Failed to parse Windows release id: %s", stdOut)
	}

	if yourWindowsReleaseId < minimumWindowsReleaseId {
		return errors.Newf("Please update Windows. Currently %d is the minimum release needed to run. You are running %d", minimumWindowsReleaseId, yourWindowsReleaseId)
	}
	return nil
}

func checkHyperVInstalled() error {
	// check to see if a hypervisor is present. if hyper-v is installed and enabled,
	checkHypervisorPresent := `@(Get-Wmiobject Win32_ComputerSystem).HypervisorPresent`
	stdOut, _, err := powershell.Execute(checkHypervisorPresent)
	if err != nil {
		logging.Debug(err.Error())
		return errors.New("Failed checking if Hyper-V is installed")
	}
	if !strings.Contains(stdOut, "True") {
		return errors.New("Hyper-V not installed")
	}

	checkVmmsExists := `@(Get-Service vmms).Status`
	_, stdErr, err := powershell.Execute(checkVmmsExists)
	if err != nil {
		logging.Debug(err.Error())
		return errors.New("Failed checking if Hyper-V management service exists")
	}
	if strings.Contains(stdErr, "Get-Service") {
		return errors.New("Hyper-V management service not available")
	}

	return nil
}

//
func fixHyperVInstalled() error {
	enableHyperVCommand := `Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -All`
	_, _, err := powershell.ExecuteAsAdmin("enable Hyper-V", enableHyperVCommand)

	if err != nil {
		logging.Debug(err.Error())
		return errors.New("Error occurred installing Hyper-V")
	}

	// We do need to error out as a restart might be needed (unfortunately no output redirect possible)
	logging.Error("Please reboot your system")
	return nil
}

func checkHyperVServiceRunning() error {
	// Check if Hyper-V's Virtual Machine Management Service is running
	checkVmmsRunning := `@(Get-Service vmms).Status`
	stdOut, _, err := powershell.Execute(checkVmmsRunning)
	if err != nil {
		logging.Debug(err.Error())
		return errors.New("Failed checking if Hyper-V is running")
	}
	if strings.TrimSpace(stdOut) != "Running" {
		return errors.New("Hyper-V Virtual Machine Management service not running")
	}

	return nil
}

func fixHyperVServiceRunning() error {
	enableVmmsService := `Set-Service -Name vmms -StartupType Automatic; Set-Service -Name vmms -Status Running -PassThru`
	_, _, err := powershell.ExecuteAsAdmin("enable Hyper-V service", enableVmmsService)

	if err != nil {
		logging.Debug(err.Error())
		return errors.New("Error occurred enabling Hyper-V service")
	}

	return nil
}

func checkIfUserPartOfHyperVAdmins() error {
	// https://support.microsoft.com/en-us/help/243330/well-known-security-identifiers-in-windows-operating-systems
	// BUILTIN\Hyper-V Administrators => S-1-5-32-578

	checkIfMemberOfHyperVAdmins :=
		`$sid = New-Object System.Security.Principal.SecurityIdentifier("S-1-5-32-578")
	@([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole($sid)`
	stdOut, _, err := powershell.Execute(checkIfMemberOfHyperVAdmins)
	if err != nil {
		logging.Debug(err.Error())
		return errors.New("Failed checking if user is part of hyperv admins group")
	}
	if !strings.Contains(stdOut, "True") {
		return errors.New("User is not a member of the Hyper-V administrators group")
	}

	return nil
}

func fixUserPartOfHyperVAdmins() error {
	outGroupName, _, err := powershell.Execute(`(New-Object System.Security.Principal.SecurityIdentifier("S-1-5-32-578")).Translate([System.Security.Principal.NTAccount]).Value`)
	if err != nil {
		logging.Debug(err.Error())
		return errors.New("Unable to get group name")
	}
	groupName := strings.TrimSpace(strings.Replace(strings.TrimSpace(outGroupName), "BUILTIN\\", "", -1))

	username := os.Getenv("USERNAME")

	netCmdArgs := fmt.Sprintf(`([adsi]"WinNT://./%s,group").Add("WinNT://%s,user")`, groupName, username)
	_, _, err = powershell.ExecuteAsAdmin("add user to hyperv admins group", netCmdArgs)
	if err != nil {
		logging.Debug(err.Error())
		return errors.New("Unable to get user name")
	}

	return nil
}

func checkIfHyperVVirtualSwitchExists() error {
	switchName := hyperv.AlternativeNetwork

	// use winnet instead
	exists, foundName := winnet.SelectSwitchByNameOrDefault(switchName)
	if exists {
		logging.Info("Found Virtual Switch to use: ", foundName)
		return nil
	}

	return errors.New("Virtual Switch not found")
}

func checkIfRunningAsNormalUser() error {
	if !powershell.IsAdmin() {
		return nil
	}
	logging.Debug("Ran as administrator")
	return fmt.Errorf("crc should be ran as a normal user")
}
