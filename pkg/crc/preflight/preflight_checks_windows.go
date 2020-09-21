package preflight

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/code-ready/crc/pkg/crc/logging"

	winnet "github.com/code-ready/crc/pkg/os/windows/network"
	"github.com/code-ready/crc/pkg/os/windows/powershell"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/hyperv"
)

const (
	// Fall Creators update comes with the "Default Switch"
	minimumWindowsReleaseID = 1709
)

func checkVersionOfWindowsUpdate() error {
	windowsReleaseID := `(Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion" -Name ReleaseId).ReleaseId`

	stdOut, _, err := powershell.Execute(windowsReleaseID)
	if err != nil {
		logging.Debug(err.Error())
		return fmt.Errorf("Failed to get Windows release id")
	}
	yourWindowsReleaseID, err := strconv.Atoi(strings.TrimSpace(stdOut))

	if err != nil {
		logging.Debug(err.Error())
		return fmt.Errorf("Failed to parse Windows release id: %s", stdOut)
	}

	if yourWindowsReleaseID < minimumWindowsReleaseID {
		return fmt.Errorf("Please update Windows. Currently %d is the minimum release needed to run. You are running %d", minimumWindowsReleaseID, yourWindowsReleaseID)
	}
	return nil
}

func checkWindowsEdition() error {
	windowsEditionCmd := `(Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion").EditionID`

	stdOut, _, err := powershell.Execute(windowsEditionCmd)
	if err != nil {
		logging.Debug(err.Error())
		return fmt.Errorf("Failed to get Windows edition")
	}

	windowsEdition := strings.TrimSpace(stdOut)
	logging.Debugf("Running on Windows %s edition", windowsEdition)

	if strings.HasPrefix(windowsEdition, "Home") {
		return fmt.Errorf("Windows Home edition is not supported")
	}

	return nil
}

func checkHyperVInstalled() error {
	// check to see if a hypervisor is present. if hyper-v is installed and enabled,
	checkHypervisorPresent := `@(Get-Wmiobject Win32_ComputerSystem).HypervisorPresent`
	stdOut, _, err := powershell.Execute(checkHypervisorPresent)
	if err != nil {
		logging.Debug(err.Error())
		return fmt.Errorf("Failed checking if Hyper-V is installed")
	}
	if !strings.Contains(stdOut, "True") {
		return fmt.Errorf("Hyper-V not installed")
	}

	checkVmmsExists := `@(Get-Service vmms).Status`
	_, stdErr, err := powershell.Execute(checkVmmsExists)
	if err != nil {
		logging.Debug(err.Error())
		return fmt.Errorf("Failed checking if Hyper-V management service exists")
	}
	if strings.Contains(stdErr, "Get-Service") {
		return fmt.Errorf("Hyper-V management service not available")
	}

	return nil
}

//
func fixHyperVInstalled() error {
	enableHyperVCommand := `Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -All`
	_, _, err := powershell.ExecuteAsAdmin("enable Hyper-V", enableHyperVCommand)

	if err != nil {
		logging.Debug(err.Error())
		return fmt.Errorf("Error occurred installing Hyper-V")
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
		return fmt.Errorf("Failed checking if Hyper-V is running")
	}
	if strings.TrimSpace(stdOut) != "Running" {
		return fmt.Errorf("Hyper-V Virtual Machine Management service not running")
	}

	return nil
}

func fixHyperVServiceRunning() error {
	enableVmmsService := `Set-Service -Name vmms -StartupType Automatic; Set-Service -Name vmms -Status Running -PassThru`
	_, _, err := powershell.ExecuteAsAdmin("enable Hyper-V service", enableVmmsService)

	if err != nil {
		logging.Debug(err.Error())
		return fmt.Errorf("Error occurred enabling Hyper-V service")
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
		return fmt.Errorf("Failed checking if user is part of hyperv admins group")
	}
	if !strings.Contains(stdOut, "True") {
		return fmt.Errorf("User is not a member of the Hyper-V administrators group")
	}

	return nil
}

func fixUserPartOfHyperVAdmins() error {
	outGroupName, _, err := powershell.Execute(`(New-Object System.Security.Principal.SecurityIdentifier("S-1-5-32-578")).Translate([System.Security.Principal.NTAccount]).Value`)
	if err != nil {
		logging.Debug(err.Error())
		return fmt.Errorf("Unable to get group name")
	}
	groupName := strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(outGroupName), "BUILTIN\\", ""))

	username := os.Getenv("USERNAME")

	netCmdArgs := fmt.Sprintf(`([adsi]"WinNT://./%s,group").Add("WinNT://%s,user")`, groupName, username)
	_, _, err = powershell.ExecuteAsAdmin("add user to hyperv admins group", netCmdArgs)
	if err != nil {
		logging.Debug(err.Error())
		return fmt.Errorf("Unable to get user name")
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

	return fmt.Errorf("Virtual Switch not found")
}

func checkIfRunningAsNormalUser() error {
	if !powershell.IsAdmin() {
		return nil
	}
	logging.Debug("Ran as administrator")
	return fmt.Errorf("crc should be ran as a normal user")
}

func removeDNSServerAddress() error {
	resetDNSCommand := `Set-DnsClientServerAddress -InterfaceAlias ("vEthernet (crc)") -ResetServerAddresses`
	if exist, defaultSwitch := winnet.GetDefaultSwitchName(); exist {
		resetDNSCommand = fmt.Sprintf(`Set-DnsClientServerAddress -InterfaceAlias ("vEthernet (%s)","vEthernet (crc)") -ResetServerAddresses`, defaultSwitch)
	}
	if _, _, err := powershell.ExecuteAsAdmin("Remove dns entry for default switch", resetDNSCommand); err != nil {
		return err
	}
	return nil
}

func removeCrcVM() (err error) {
	defer func() {
		if ferr := os.RemoveAll(constants.MachineInstanceDir); ferr != nil {
			logging.Debugf("Error removing %s dir: %v", constants.MachineInstanceDir, err)
			err = fmt.Errorf("Error removing %s dir", constants.MachineInstanceDir)
		}
	}()
	if _, _, err := powershell.Execute("Get-VM -Name crc"); err != nil {
		// This means that there is no crc VM exist
		return nil
	}
	stopVMCommand := fmt.Sprintf(`Stop-VM -Name "%s" -Force`, constants.DefaultName)
	if _, _, err := powershell.Execute(stopVMCommand); err != nil {
		// ignore the error as this is useless (prefer not to use nolint here)
		return err
	}
	removeVMCommand := fmt.Sprintf(`Remove-VM -Name "%s" -Force`, constants.DefaultName)
	if _, _, err := powershell.Execute(removeVMCommand); err != nil {
		// ignore the error as this is useless (prefer not to use nolint here)
		return err
	}
	logging.Debug("'crc' VM is removed")
	return nil
}

func setSuid(path string) error {
	return nil
}

func checkSuid(path string) error {
	return nil
}
