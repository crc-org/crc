package hyperv

import (
	"errors"
	"fmt"

	log "github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/os/windows/powershell"
	crcstrings "github.com/crc-org/crc/v2/pkg/strings"
)

var (
	ErrPowerShellNotFound = errors.New("Powershell was not found in the path")
	ErrNotAdministrator   = errors.New("Hyper-v commands have to be run as an Administrator")
	ErrNotInstalled       = errors.New("Hyper-V PowerShell Module is not available")
)

func cmdOut(args ...string) (string, error) {
	stdout, _, err := powershell.Execute(args...)
	return stdout, err
}

func cmd(args ...string) error {
	_, err := cmdOut(args...)
	return err
}

func hypervAvailable() error {
	stdout, err := cmdOut("@(Get-Module -ListAvailable hyper-v).Name | Get-Unique")
	if err != nil {
		return err
	}

	resp := crcstrings.FirstLine(stdout)
	if resp != "Hyper-V" {
		return ErrNotInstalled
	}

	return nil
}

func isAdministrator() (bool, error) {
	hypervAdmin := isHypervAdministrator()

	if hypervAdmin {
		return true, nil
	}

	windowsAdmin, err := isWindowsAdministrator()

	if err != nil {
		return false, err
	}

	return windowsAdmin, nil
}

func isHypervAdministrator() bool {
	stdout, err := cmdOut(`@([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole(([System.Security.Principal.SecurityIdentifier]::new("S-1-5-32-578")))`)
	if err != nil {
		log.Debug(err)
		return false
	}

	resp := crcstrings.FirstLine(stdout)
	return resp == "True"
}

func isWindowsAdministrator() (bool, error) {
	stdout, err := cmdOut(`@([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")`)
	if err != nil {
		return false, err
	}

	resp := crcstrings.FirstLine(stdout)
	return resp == "True", nil
}

func quote(text string) string {
	return fmt.Sprintf("'%s'", text)
}

func toMb(value int) string {
	return fmt.Sprintf("%dMB", value)
}

func smbShareExists(name string) bool {
	if err := cmd(fmt.Sprintf("Get-SmbShare -Name %s", name)); err != nil {
		return false
	}
	return true
}
