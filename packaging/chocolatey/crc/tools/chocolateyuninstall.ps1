$ErrorActionPreference = 'SilentlyContinue';
$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$setupScript = Join-Path "$toolsDir" "crcprerequisitesetup.ps1"

Import-Module $setupScript

# remove crc-users group
Remove-LocalGroup -Name 'crc-users' -Confirm:$false | Out-Null
