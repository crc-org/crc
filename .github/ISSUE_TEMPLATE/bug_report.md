---
name: "Bug report"
about: "Create a report to help us improve"
title: "[BUG]"
labels: "kind/bug, status/need triage"
assignees: ""

---

### General information

  * OS: Linux / macOS / Windows
  * Hypervisor: KVM / Hyper-V / hyperkit
  * Did you run `crc setup` before starting it (Yes/No)?
  * Running CRC on: Laptop / Baremetal-Server / VM

## CRC version
```bash
# Put `crc version` output here
```
  
## CRC status
```bash
# Put `crc status --log-level debug` output here
```

## CRC config
```bash
# Put `crc config view` output here
```

## Host Operating System
```bash
# Put the output of `cat /etc/os-release` in case of Linux
# put the output of `sw_vers` in case of Mac
# Put the output of `systeminfo` in case of Windows
```

### Steps to reproduce

  1. 
  2. 
  3. 
  4. 

### Expected


### Actual


### Logs

Before gather the logs try following if that fix your issue
```bash
$ crc delete -f
$ crc cleanup
$ crc setup
$ crc start --log-level debug
```

Please consider posting the output of `crc start --log-level debug`  on http://gist.github.com/ and post the link in the issue.
