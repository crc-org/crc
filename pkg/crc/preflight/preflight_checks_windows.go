package preflight

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/logging"

	"github.com/crc-org/crc/v2/pkg/crc/adminhelper"
	winnet "github.com/crc-org/crc/v2/pkg/os/windows/network"
	"github.com/crc-org/crc/v2/pkg/os/windows/powershell"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/machine/hyperv"
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

	if strings.ToLower(windowsEdition) == "core" {
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

func checkUserPartOfCrcUsersAndHypervAdminsGroup() error {
	groupMembers, _, err := powershell.Execute(`(Get-LocalGroupMember -Group 'crc-users').Name`)
	if err != nil {
		return err
	}
	logging.Debug("Checking current user is in the 'crc-user' group")
	if !usernameInMembersList(username(), groupMembers) {
		return fmt.Errorf("Could not find: %s in the 'crc-users' group", username())
	}
	// https://support.microsoft.com/en-us/help/243330/well-known-security-identifiers-in-windows-operating-systems
	// BUILTIN\Hyper-V Administrators => S-1-5-32-578
	groupMembers, _, err = powershell.Execute(`(Get-LocalGroupMember -SID 'S-1-5-32-578').Name`)
	if err != nil {
		return err
	}
	logging.Debug("Checking current user is in the 'Hyper-v Administrators' group")
	if !usernameInMembersList(username(), groupMembers) {
		return fmt.Errorf("Could not find: %s in the 'Hyper-v Administrators' group", username())
	}
	return nil
}

func usernameInMembersList(username, members string) bool {
	m := strings.Split(members, "\n")
	var memberList []string
	// remove any empty elements
	for _, elem := range m {
		if strings.TrimSpace(elem) != "" {
			memberList = append(memberList, strings.TrimSpace(elem))
		}
	}
	logging.Debugf("group members: %s", strings.Join(memberList, ","))
	for _, member := range memberList {
		// we get the members of a group in the form domain\username
		// if the current username is also returned in the same form
		// check if that full domain\username is present in the list
		if strings.Contains(username, "\\") {
			if username == member {
				return true
			}
			continue
		}

		// if we get only the username without domain part then we
		// compare it with the username part of the group members
		m := strings.Split(member, "\\")
		if len(m) == 2 {
			if m[1] == username {
				return true
			}
		} else {
			logging.Warnf("Got group member's name in unexpected format: %s", member)
		}
	}
	return false
}

func fixUserPartOfCrcUsersAndHypervAdminsGroup() error {
	cmd := `Add-LocalGroupMember -Group 'crc-users' -Member '%s';Add-LocalGroupMember -SID 'S-1-5-32-578' -Member '%s'`
	_, _, err := powershell.ExecuteAsAdmin("adding current user to crc-users and Hyper-V admins group", fmt.Sprintf(cmd, username(), username()))
	if err != nil {
		return err
	}
	return errReboot
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
	return fmt.Errorf("crc should be ran in a shell without administrator rights")
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
	if _, _, err := powershell.Execute("Get-VM -Name crc"); err != nil {
		// This means that there is no crc VM exist
		return nil
	}
	stopVMCommand := fmt.Sprintf(`Stop-VM -Name "%s" -TurnOff -Force`, constants.DefaultName)
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

func checkIfAdminHelperServiceRunning() error {
	stdout, stderr, err := powershell.Execute(fmt.Sprintf("(Get-Service %s).Status", constants.AdminHelperServiceName))
	if err != nil {
		return fmt.Errorf("%s service is not present %v: %s", constants.AdminHelperServiceName, err, stderr)
	}
	if strings.TrimSpace(stdout) != "Running" {
		return fmt.Errorf("%s service is not running", constants.AdminHelperServiceName)
	}
	return checkAdminHelperNamedPipeAccessible()
}

func checkAdminHelperNamedPipeAccessible() error {
	client := adminhelper.Client()
	if _, err := client.Version(); err != nil {
		return err
	}
	return nil
}
