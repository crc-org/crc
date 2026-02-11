# Testing Rosetta Support

## Prerequisites

- Apple Silicon Mac (M1/M2/M3/M4)
- macOS 13+ (Ventura or later)
- CRC built from `feat/rosetta-support` branch

## Quick Start

```bash
# 1. Enable Rosetta
crc config set use-rosetta true

# 2. Delete existing VM (required — Rosetta device must be present at VM creation)
crc delete -f

# 3. Setup and start
crc setup
crc start
```

## Verify Rosetta Is Working Inside the VM

SSH into the VM and check:

```bash
# SSH in
crc ssh

# Check that Rosetta binfmt is registered
cat /proc/sys/fs/binfmt_misc/rosetta
# Should show:
#   enabled
#   interpreter /media/rosetta/rosetta
#   flags: CF

# Confirm Rosetta mount exists
mount | grep rosetta
# Should show: rosetta on /media/rosetta type virtiofs (...)

# Confirm QEMU x86_64 handler is disabled/absent
cat /proc/sys/fs/binfmt_misc/qemu-x86_64 2>/dev/null || echo "qemu-x86_64 not registered (good)"
# Should show "disabled" or "not registered"
```

## Test With a Known x86_64 Container That Fails Under QEMU

The classic failure case is Go binaries crashing with `lfstack.push` panics under QEMU. These containers are x86_64-only and will crash under QEMU but run fine with Rosetta:

```bash
# SSH in first
crc ssh

# Test 1: Run an x86_64 Go binary (etcd) — crashes under QEMU, works with Rosetta
sudo podman run --rm --platform linux/amd64 quay.io/coreos/etcd:v3.5.9 etcd --version

# Test 2: Run an x86_64 .NET container — known to be slow/broken under QEMU
sudo podman run --rm --platform linux/amd64 mcr.microsoft.com/dotnet/runtime:8.0 dotnet --info

# Test 3: Simple x86_64 container to confirm basic translation works
sudo podman run --rm --platform linux/amd64 docker.io/library/alpine:latest uname -m
# Should print: x86_64
```

## Test Without Rosetta (Baseline / Reproduce QEMU Failures)

```bash
crc config set use-rosetta false
crc delete -f
crc setup
crc start
crc ssh

# This will likely crash or hang under QEMU on Apple Silicon:
sudo podman run --rm --platform linux/amd64 quay.io/coreos/etcd:v3.5.9 etcd --version
```

## What to Look For

### Success indicators
- `crc setup` shows "Checking if Rosetta is installed" preflight check passing
- `crc start` logs "Configuring Rosetta for x86_64 emulation"
- `/proc/sys/fs/binfmt_misc/rosetta` exists and shows `enabled`
- x86_64 containers run without `lfstack.push` panics or segfaults
- Go-based x86_64 binaries (etcd, kubectl, etc.) execute correctly

### Failure indicators
- `crc setup` fails at Rosetta preflight — run `softwareupdate --install-rosetta` manually
- `crc start` errors with "Failed to configure Rosetta" — check SSH connectivity, check if mount failed
- Rosetta binfmt not registered — SSH in and manually check `/proc/sys/fs/binfmt_misc/`
- SELinux denials — check `sudo ausearch -m avc -ts recent` inside the VM
- Containers still crashing — verify `qemu-x86_64` handler is actually disabled

### Debugging

```bash
# Inside the VM, check all registered binfmt handlers
ls /proc/sys/fs/binfmt_misc/

# Check SELinux context on Rosetta mount
ls -laZ /media/rosetta/

# Check if Rosetta binary is accessible
file /media/rosetta/rosetta

# Check systemd-binfmt service status
systemctl status systemd-binfmt

# Look at journal for errors during start
journalctl -u systemd-binfmt --no-pager
```

## Config Validation

```bash
# Should succeed on Apple Silicon Mac
crc config set use-rosetta true

# Should fail on Intel Mac or Linux
crc config set use-rosetta true
# Expected: "Rosetta is only supported on Apple Silicon Macs"

# Toggling requires VM delete
crc config set use-rosetta false
# Expected message: requires delete and setup
```
