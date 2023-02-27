$ErrorActionPreference = 'Stop';

Stop-Service -Name 'crcAdminHelper' -Force -Confirm:$false -ErrorAction SilentlyContinue | Out-Null
sc.exe delete "crcAdminHelper" | Out-Null

Stop-ScheduledTask -TaskName 'crcDaemon' -ErrorAction SilentlyContinue | Out-Null
Unregister-ScheduledTask -TaskName 'crcDaemon' -Confirm:$false -ErrorAction SilentlyContinue | Out-Null
