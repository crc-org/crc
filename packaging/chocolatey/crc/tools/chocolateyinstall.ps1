﻿$ErrorActionPreference = 'Stop'; # stop on all errors
$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$setupScript = Join-Path "$toolsDir" "crcprerequisitesetup.ps1"

Import-Module $setupScript

# generate ignore file for crc-admin-helper-windows.exe to avoid shim generation and addintion to PATH
New-Item "$toolsDir\crc-admin-helper-windows.exe.ignore" -ItemType File -Force | Out-Null

if (Test-ProcessAdminRights) {
    New-CrcGroup
    Install-Hyperv
    Install-AdminHelper -AdminHelperPath "$toolsDir\crc-admin-helper-windows.exe"
    New-VsockGuestCommunicationRegistry
} else {
    Write-Host "CRC needs administrator privileges to enable Hyper-V, create new group and add current user to needed groups, please run the installation as administrator"
    Exit 1
}
