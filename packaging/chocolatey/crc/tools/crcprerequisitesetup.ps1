$CrcGroupName = 'crc-users'
$CrcAdminHelperServiceName = 'crcAdminHelper'
$VsockGuestCommunicationRegistryPath = "HKLM:\Software\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices\00000400-FACB-11E6-BD58-64006A7986D3"

function New-CrcGroup {
    $crcgrp = Get-LocalGroup -Name $CrcGroupName -ErrorAction SilentlyContinue
    if ($crcgrp) {
        Write-Host "crc-users group already exists"
    } else {
        New-LocalGroup -Name $CrcGroupName -Description 'Group for Red Hat OpenShift Local users' | Out-Null
    }
}

function Install-Hyperv {
    $osProductType = Get-ComputerInfo | Select-Object -ExpandProperty OSProductType | Out-String -Stream | Where-Object { $_.Trim().Length -gt 0 }
    switch ($osProductType)
    {
        "WorkStation" {
            # check if hyperv is already enabled
            $enabled = (Get-WindowsOptionalFeature -FeatureName "Microsoft-Hyper-V" -online | Select-Object -ExpandProperty State).ToString().Equals("Enabled")
            if ($enabled) {
                Write-Host "Hyper-V is already enabled"
            } else {
                Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -All -NoRestart -ErrorAction SilentlyContinue | Out-Null
            }
        }
        "Server" {
            $enabled = (Get-WindowsFeature -Name "Hyper-V" | Select-Object -ExpandProperty InstallState).ToString().Equals("Installed")
            if ($enabled) {
                Write-Host "Hyper-V is already enabled"
            } else {
                Install-WindowsFeature -Name Hyper-V -IncludeManagementTools -ErrorAction SilentlyContinue | Out-Null
            }
        }
    }
}

function Install-AdminHelper([string]$AdminHelperPath) {
    $svc = Get-Service -Name "$CrcAdminHelperServiceName" -ErrorAction SilentlyContinue
    if ($svc) {
        Write-Host "Red Hat OpenShift Local Admin Helper service is already installed"
    } else {
        # New-Service cmdlet doesn't have a flag to set the arguments for admin-helper, it is passed using the -BinaryPathName itself
        New-Service -Name 'crcAdminHelper' -DisplayName "Red Hat OpenShift Local Admin Helper" -Description "Perform administrative tasks for the user" -StartupType Automatic -BinaryPathName "$AdminHelperPath daemon" | Out-Null
    }

    Start-Service -Name 'crcAdminHelper' -Confirm:$false | Out-Null
}

function New-VsockGuestCommunicationRegistry {
    # check if the registry path exists
    $pathExists = Test-Path -Path $VsockGuestCommunicationRegistryPath -ErrorAction SilentlyContinue
    # check if the property 'ElementName' exists at the path
    $property = Get-ItemProperty -Path $VsockGuestCommunicationRegistryPath -Name 'ElementName' -ErrorAction SilentlyContinue
    $propertyExists = $property.ElementName -eq "gvisor-tap-vsock"
    if ($pathExists -and $propertyExists) {
        Write-Host "Property ElementName in $VsockGuestCommunicationRegistryPath already exists"
    } else {
        # create the registry node
        New-Item -Path $VsockGuestCommunicationRegistryPath -ErrorAction SilentlyContinue | Out-Null
        # remove any pre-existing value
        Remove-ItemProperty -Path $VsockGuestCommunicationRegistryPath -Name 'ElementName' -ErrorAction SilentlyContinue | Out-Null
        # add the 'ElementName' property
        New-ItemProperty -Path $VsockGuestCommunicationRegistryPath -Name 'ElementName' -Value 'gvisor-tap-vsock' -PropertyType string -Confirm:$false | Out-Null
    }
}
