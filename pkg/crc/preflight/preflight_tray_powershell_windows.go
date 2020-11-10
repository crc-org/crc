package preflight

import (
	"fmt"
	"strings"
)

var (
	trayInstallationScript = []string{
		`$ErrorActionPreference = "Stop"`,
		`$password = "%s"`,
		`$tempDir = "%s"`,
		`$crcExecutablePath = "%s"`,
		`$trayExecutablePath = "%s"`,
		`$traySymlinkName = "%s"`,
		`$serviceName = "%s"`,
		`$currentUserSid = (Get-LocalUser -Name "$env:USERNAME").Sid.Value`,
		`$startUpFolder = "$Env:USERPROFILE\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup"`,
		`function AddServiceLogonRightForCurrentUser()`,
		`{`,
		`	$securityTemplate = @"`,
		`[Unicode]`,
		`Unicode=yes`,
		`[Version]`,
		"signature=\"`$CHICAGO$\"",
		`Revision=1`,
		`[Privilege Rights]`,
		`SeServiceLogonRight = {0}`,
		`"@`,
		`	SecEdit.exe /export /cfg $tempDir\secdef.inf /areas USER_RIGHTS`,
		`	if ($LASTEXITCODE -ne 0)`,
		`	{`,
		`		exit 1`,
		`	}`,
		`	$userRights = Get-Content -Path $tempDir\secdef.inf`,
		`	$serviceLogonUserRight = ($userRights | select-string -Pattern "SeServiceLogonRight\s=\s.*")`,
		`	$sidsInServiceLogonRight = ($serviceLogonUserRight -split "=")[1].Trim()`,
		`	$sidsArray = $sidsInServiceLogonRight -split ","`,
		`	if (!($sidsArray.Contains($env:USERNAME) -or $sidsArray.Contains("*"+$currentUserSid)))`,
		`	{`,
		`		Write-Output "User doesn't have logon as service right, adding sid of $env:Username"`,
		`		$sidsInServiceLogonRight += ",*$currentUserSid"`,
		`		$templateContent = $securityTemplate -f "$sidsInServiceLogonRight"`,
		`		Set-Content -Path $tempDir\secdef_fin.inf $templateContent`,
		`		SecEdit.exe /configure /db $tempDir\tempdb.db /cfg $tempDir\secdef_fin.inf /areas USER_RIGHTS`,
		`		if ($LASTEXITCODE -ne 0)`,
		`		{`,
		`			exit`,
		`		}`,
		`	}`,
		`}`,

		`function CreateDaemonService()`,
		`{`,
		`	$secPass = ConvertTo-SecureString $password -AsPlainText -Force`,
		`	$creds = New-Object pscredential ("$env:USERDOMAIN\$env:USERNAME", $secPass)`,
		`	$params = @{`,
		`		Name = "$serviceName"`,
		`		BinaryPathName = "$crcExecutablePath daemon"`,
		`		DisplayName = "$serviceName"`,
		`		StartupType = "Automatic"`,
		`		Description = "CodeReady Containers Daemon service for System Tray."`,
		`		Credential = $creds`,
		`	}`,
		`	New-Service @params`,
		`}`,

		`function StartDaemonService()`,
		`{`,
		`	Start-Service "CodeReady Containers"`,
		`}`,

		`AddServiceLogonRightForCurrentUser`,
		`sc.exe stop "$serviceName"`,
		`sc.exe delete "$serviceName"`,
		`CreateDaemonService`,
		`StartDaemonService`,
		`$ErrorActionPreference = "Continue"`,
		`Stop-Process -Name tray-windows`,
		`Remove-Item "$startUpFolder\$traySymlinkName"`,

		`$ErrorActionPreference = "Stop"`,
		`New-Item -ItemType SymbolicLink -Path "$startUpFolder" -Name "$traySymlinkName" -Value "$trayExecutablePath"`,
		`Start-Process -FilePath "$trayExecutablePath"`,
		`New-Item -ItemType File -Path "$tempDir" -Name "success"`,
		`Set-Content -Path $tempDir\success "blah blah"`,
	}

	trayRemovalScript = []string{
		`$tempDir = "%s"`,
		`$trayProcessName = "%s"`,
		`$traySymlinkName = "%s"`,
		`$serviceName = "%s"`,
		`$startUpFolder = "$Env:USERPROFILE\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup"`,

		`function RemoveUserFromServiceLogon`,
		`{`,
		`	$securityTemplate = @"`,
		`[Unicode]`,
		`Unicode=yes`,
		`[Version]`,
		"signature=\"`$CHICAGO$\"",
		`Revision=1`,
		`[Privilege Rights]`,
		`SeServiceLogonRight = {0}`,
		`"@`,
		`	SecEdit.exe /export /cfg $tempDir\secdef.inf /areas USER_RIGHTS`,
		`	if ($LASTEXITCODE -ne 0)`,
		`	{`,
		`		exit 1`,
		`	}`,

		`	$userRights = Get-Content -Path $tempDir\secdef.inf`,
		`	$serviceLogonUserRight = ($userRights | select-string -Pattern "SeServiceLogonRight\s=\s.*")`,

		`	$sidsInServiceLogonRight = ($serviceLogonUserRight -split "=")[1].Trim()`,
		`	$sidsArray = $sidsInServiceLogonRight -split ","`,
		`	$newSids = $sidsArray | Where-Object {$_ -ne $env:USERNAME}`,
		`	$newSids = $newSids -Join ","`,
		`	$templateContent = $securityTemplate -f "$newSids"`,

		`	Set-Content -Path $tempDir\secdef_fin.inf $templateContent`,
		`	SecEdit.exe /configure /db $tempDir\tempdb.db /cfg $tempDir\secdef_fin.inf /areas USER_RIGHTS`,
		`}`,

		`function DeleteDaemonService()`,
		`{`,
		`	sc.exe stop "$serviceName"`,
		`	sc.exe delete "$serviceName"`,
		`}`,

		`function RemoveTrayFromStartUpFolder()`,
		`{`,
		`	Stop-Process -Name "$trayProcessName"`,
		`	Remove-Item "$startUpFolder\$traySymlinkName"`,
		`}`,

		`RemoveUserFromServiceLogon`,
		`DeleteDaemonService`,
		`RemoveTrayFromStartUpFolder`,
	}
)

func getTrayInstallationScriptTemplate() string {
	return strings.Join(trayInstallationScript, "\n")
}

func getTrayRemovalScriptTemplate() string {
	return strings.Join(trayRemovalScript, "\n")
}

func genTrayInstallScript(password, tempDirPath, daemonCmd, trayExecutablePath, traySymlinkName, daemonServiceName string) string {
	return fmt.Sprintf(getTrayInstallationScriptTemplate(),
		password,
		tempDirPath,
		daemonCmd,
		trayExecutablePath,
		traySymlinkName,
		daemonServiceName,
	)
}

func genTrayRemovalScript(trayProcessName, traySymlinkName, daemonServiceName, tempDir string) string {
	return fmt.Sprintf(getTrayRemovalScriptTemplate(),
		tempDir,
		trayProcessName,
		traySymlinkName,
		daemonServiceName,
	)
}
