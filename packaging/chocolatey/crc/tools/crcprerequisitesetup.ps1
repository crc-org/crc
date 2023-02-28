$ErrorActionPreference = 'Stop';
$CrcGroupName = 'crc-users'
$CrcAdminHelperServiceName = 'crcAdminHelper'
$VsockGuestCommunicationRegistryPath = "HKLM:\Software\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices\00000400-FACB-11E6-BD58-64006A7986D3"

function New-CrcGroup {
    $crcgrp = Get-LocalGroup -Name $CrcGroupName -ErrorAction SilentlyContinue
    if ($crcgrp) {
        Write-Host "Local Group 'crc-users' already exists"
    } else {
        New-LocalGroup -Name $CrcGroupName -Description 'Group for Red Hat OpenShift Local users' | Out-Null
    }
}

function Install-Hyperv {
    $osProductType = Get-ComputerInfo | Select-Object -ExpandProperty OSProductType | Out-String -Stream | Where-Object { $_.Trim().Length -gt 0 }
    switch ($osProductType)
    {
        "WorkStation" {
            $enabled = (Get-WindowsOptionalFeature -FeatureName "Microsoft-Hyper-V" -online | Select-Object -ExpandProperty State).ToString().Equals("Enabled")
            if ($enabled) {
                Write-Host "Hyper-V is already enabled"
            } else {
                try {
                    Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -All -NoRestart | Out-Null
                    Write-Host -ForegroundColor Red "Hyper-V has been enabled, please reboot to complete installation"
                }
                catch {
                    Write-Host -ForegroundColor Red "Failed to enable Hyper-V, aborting installation"
                    Set-PowershellExitCode 1
                }
            }
        }
        "Server" {
            $enabled = (Get-WindowsFeature -Name "Hyper-V" | Select-Object -ExpandProperty InstallState).ToString().Equals("Installed")
            if ($enabled) {
                Write-Host "Hyper-V is already enabled"
            } else {
                try {
                    Install-WindowsFeature -Name Hyper-V -IncludeManagementTools -Confirm:$false -Restart:$false | Out-Null
                    Write-Host -ForegroundColor Red "Hyper-V has been enabled, please reboot to complete installation"
                }
                catch {
                    Write-Host -ForegroundColor Red "Failed to enable Hyper-V, aborting installation"
                    Set-PowershellExitCode 1
                }
            }
        }
    }
}

function Install-AdminHelper([string]$AdminHelperPath) {
    $svc = Get-Service -Name "$CrcAdminHelperServiceName" -ErrorAction SilentlyContinue
    if ($svc) {
        Write-Host "Red Hat OpenShift Local Admin Helper service is already installed"
    } else {
        New-Service -Name 'crcAdminHelper' -DisplayName "Red Hat OpenShift Local Admin Helper" -Description "Perform administrative tasks for the user" -StartupType Automatic -BinaryPathName "$AdminHelperPath daemon" | Out-Null
    }
    Start-Service -Name 'crcAdminHelper' -Confirm:$false -ErrorAction SilentlyContinue | Out-Null
}

function New-VsockGuestCommunicationRegistry {
    $pathExists = Test-Path -Path $VsockGuestCommunicationRegistryPath -ErrorAction SilentlyContinue
    $property = Get-ItemProperty -Path $VsockGuestCommunicationRegistryPath -Name 'ElementName' -ErrorAction SilentlyContinue
    $propertyExists = $property.ElementName -eq "gvisor-tap-vsock"
    if ($pathExists -and $propertyExists) {
        Write-Host "Property ElementName in $VsockGuestCommunicationRegistryPath already exists"
    } else {
        New-Item -Path $VsockGuestCommunicationRegistryPath -ErrorAction SilentlyContinue | Out-Null
        Remove-ItemProperty -Path $VsockGuestCommunicationRegistryPath -Name 'ElementName' -ErrorAction SilentlyContinue | Out-Null
        New-ItemProperty -Path $VsockGuestCommunicationRegistryPath -Name 'ElementName' -Value 'gvisor-tap-vsock' -PropertyType string -Confirm:$false | Out-Null
    }
}
