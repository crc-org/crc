$ErrorActionPreference = 'Stop';
$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$setupScript = Join-Path "$toolsDir" "crcprerequisitesetup.ps1"

Import-Module $setupScript
Remove-LocalGroup -Name 'crc-users' -Confirm:$false -ErrorAction SilentlyContinue | Out-Null
