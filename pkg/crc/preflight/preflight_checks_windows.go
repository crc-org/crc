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

func checkVersionOfWindowsUpdate() (bool, error) {
	windowsReleaseId := `(Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion" -Name ReleaseId).ReleaseId`

	stdOut, _, err := powershell.Execute(windowsReleaseId)
	if err != nil {
		logging.Debug(err.Error())
		return false, errors.New("Failed to get Windows release id")
	}
	yourWindowsReleaseId, err := strconv.Atoi(strings.TrimSpace(stdOut))

	if err != nil {
		logging.Debug(err.Error())
		return false, errors.Newf("Failed to parse Windows release id: %s", stdOut)
	}

	if yourWindowsReleaseId < minimumWindowsReleaseId {
		return false, errors.Newf("Please update Windows. Currently %d is the minimum release needed to run. You are running %d", minimumWindowsReleaseId, yourWindowsReleaseId)
	}
	return true, nil
}

// Unable to update automatically
func fixVersionOfWindowsUpdate() (bool, error) {
	return false, errors.New("Please manually update your Windows 10 installation")
}

func checkHyperVInstalled() (bool, error) {
	// check to see if a hypervisor is present. if hyper-v is installed and enabled,
	checkHypervisorPresent := `@(Get-Wmiobject Win32_ComputerSystem).HypervisorPresent`
	stdOut, _, err := powershell.Execute(checkHypervisorPresent)
	if err != nil {
		logging.Debug(err.Error())
		return false, errors.New("Failed checking if Hyper-V is installed")
	}
	if !strings.Contains(stdOut, "True") {
		return false, errors.New("Hyper-V not installed")
	}

	// Check if Hyper-V's Virtual Machine Management Service is running
	checkVmmsRunning := `@(Get-Service vmms).Status`
	stdOut, _, err = powershell.Execute(checkVmmsRunning)
	if err != nil {
		logging.Debug(err.Error())
		return false, errors.New("Failed checking if Hyper-V is running")
	}
	if strings.TrimSpace(stdOut) != "Running" {
		return false, errors.New("Hyper-V Virtual Machine Management service not running")
	}

	return true, nil
}

//
func fixHyperVInstalled() (bool, error) {
	enableHyperVCommand := `Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -All`
	_, _, err := powershell.ExecuteAsAdmin("enable Hyper-V", enableHyperVCommand)

	if err != nil {
		logging.Debug(err.Error())
		return false, errors.New("Error occurred installing Hyper-V")
	}

	// We do need to error out as a restart might be needed (unfortunately no output redirect possible)
	logging.Error("Please reboot your system")
	return true, nil
}

func checkIfUserPartOfHyperVAdmins() (bool, error) {
	// https://support.microsoft.com/en-us/help/243330/well-known-security-identifiers-in-windows-operating-systems
	// BUILTIN\Hyper-V Administrators => S-1-5-32-578

	checkIfMemberOfHyperVAdmins :=
		`$sid = New-Object System.Security.Principal.SecurityIdentifier("S-1-5-32-578")
	@([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole($sid)`
	stdOut, _, err := powershell.Execute(checkIfMemberOfHyperVAdmins)
	if err != nil {
		logging.Debug(err.Error())
		return false, errors.New("Failed checking if user is part of hyperv admins group")
	}
	if !strings.Contains(stdOut, "True") {
		return false, errors.New("User is not a member of the Hyper-V administrators group")
	}

	return true, nil
}

func fixUserPartOfHyperVAdmins() (bool, error) {
	outGroupName, _, err := powershell.Execute(`(New-Object System.Security.Principal.SecurityIdentifier("S-1-5-32-578")).Translate([System.Security.Principal.NTAccount]).Value`)
	if err != nil {
		logging.Debug(err.Error())
		return false, errors.New("Unable to get group name")
	}
	groupName := strings.TrimSpace(strings.Replace(strings.TrimSpace(outGroupName), "BUILTIN\\", "", -1))

	username := os.Getenv("USERNAME")

	netCmdArgs := fmt.Sprintf(`([adsi]"WinNT://./%s,group").Add("WinNT://%s,user")`, groupName, username)
	_, _, err = powershell.ExecuteAsAdmin("add user to hyperv admins group", netCmdArgs)
	if err != nil {
		logging.Debug(err.Error())
		return false, errors.New("Unable to get user name")
	}

	return true, nil
}

func checkIfHyperVVirtualSwitchExists() (bool, error) {
	switchName := hyperv.AlternativeNetwork

	// use winnet instead
	exists, foundName := winnet.SelectSwitchByNameOrDefault(switchName)
	if exists {
		logging.Info("Found Virtual Switch to use: ", foundName)
		return true, nil
	}

	return false, errors.New("Virtual Switch not found")
}

// Unable to do for now
func fixHyperVVirtualSwitch() (bool, error) {
	return false, errors.New("Unable to perform Hyper-V administrative commands. Please make sure to re-login or reboot your system")
}

func checkIfRunningAsNormalUserInWindows() (bool, error) {
	if !powershell.IsAdmin() {
		return true, nil
	}
	logging.Debug("Ran as administrator")
	return false, fmt.Errorf("crc should be ran as a normal user")
}

func fixRunAsNormalUserInWindows() (bool, error) {
	return false, fmt.Errorf("crc should be ran as a normal user")
}
