# This runs in 0.9.10+ before upgrade and uninstall.
# Use this file to do things like stop services prior to upgrade or uninstall.
# NOTE: It is an anti-pattern to call chocolateyUninstall.ps1 from here. If you
#  need to uninstall an MSI prior to upgrade, put the functionality in this
#  file without calling the uninstall script. Make it idempotent in the
#  uninstall script so that it doesn't fail when it is already uninstalled.
# NOTE: For upgrades - like the uninstall script, this script always runs from
#  the currently installed version, not from the new upgraded package version.

$ErrorActionPreference = 'SilentlyContinue';

# stop crc-admin-helper service and remove it
Stop-Service -Name 'crcAdminHelper' -Force -Confirm:$false | Out-Null
sc.exe delete "crcAdminHelper" | Out-Null

# stop crc daemon scheduled task and remove it
Stop-ScheduledTask -TaskName 'crcDaemon' | Out-Null
Unregister-ScheduledTask -TaskName 'crcDaemon' -Confirm:$false | Out-Null
