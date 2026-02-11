# PRD: Enable Rosetta 2 Support on macOS (Apple Silicon)

**Issue:** [crc-org/crc#4881](https://github.com/crc-org/crc/issues/4881)
**Author:** cpepper
**Status:** Draft

## Problem

On Apple Silicon Macs, CRC uses QEMU user-mode emulation to run x86_64/amd64 containers inside the VM. This has known bugs — Go-based binaries frequently crash with `lfstack.push` panics due to incompatibilities between QEMU and the Go runtime. Apple's Rosetta 2 translation layer is significantly faster and more compatible, but CRC doesn't expose it as an option.

## Proposed Solution

Add a `use-rosetta` boolean config setting that, when enabled on an Apple Silicon Mac:

1. Attaches a **Rosetta virtiofs share** to the vfkit VM (the vendored vfkit library already supports this via `RosettaShare`)
2. **Configures the VM** post-boot to mount the Rosetta binary and register it via `binfmt_misc`, replacing QEMU for x86_64 translation

**User flow:**
```bash
crc config set use-rosetta true
crc start   # VM boots with Rosetta enabled
```

## Architecture / Key Files

The implementation touches **7 layers** of the stack. Here's the data flow and every file involved:

### Layer 1: Config Constant & Registration

**Files:**
- `pkg/crc/config/settings.go` — Add `UseRosetta = "use-rosetta"` constant and register the setting

**What to do:**
1. Add constant: `UseRosetta = "use-rosetta"`
2. In `RegisterSettings()`, add:
   ```go
   cfg.AddSetting(UseRosetta, false, ValidateBool, RequiresDeleteAndSetupMsg,
       "Use Rosetta to run x86_64/amd64 containers on Apple Silicon (true/false, default: false)")
   ```
3. Use `RequiresDeleteAndSetupMsg` as the callback because Rosetta requires a VM recreation — the vfkit Rosetta device must be present at VM creation time.

### Layer 2: Platform Guard (Validation)

**Files:**
- `pkg/crc/config/validations.go` — Add `validateRosetta()` or use inline validation

**What to do:**
- The setting should only be valid on `darwin` + `arm64`. Create a platform-specific validation or use a build-tagged file:
  ```go
  func validateRosetta(value interface{}) (bool, string) {
      if runtime.GOARCH != "arm64" || runtime.GOOS != "darwin" {
          return false, "Rosetta is only supported on Apple Silicon Macs"
      }
      return ValidateBool(value)
  }
  ```

### Layer 3: StartConfig / MachineConfig Plumbing

**Files:**
- `pkg/crc/machine/types/types.go` — Add `EnableRosetta bool` to `StartConfig`
- `pkg/crc/machine/config/config.go` — Add `EnableRosetta bool` to `MachineConfig`
- `cmd/crc/cmd/start.go` — Wire config value into `StartConfig`
- `pkg/crc/machine/start.go` — Pass `EnableRosetta` from `StartConfig` into `MachineConfig`

### Layer 4: vfkit Driver — Add Rosetta Device to VM

**Files:**
- `pkg/drivers/vfkit/driver_darwin.go` — Add `Rosetta bool` field to `Driver` struct; in `Start()`, conditionally add `RosettaShare` device
- `pkg/crc/machine/vfkit/driver_darwin.go` — In `CreateHost()`, set `vfDriver.Rosetta` from `MachineConfig.EnableRosetta`

### Layer 5: In-VM Configuration (Post-Boot)

**Files:**
- `pkg/crc/machine/start.go` — Add `configureRosetta()` function, called during start after SSH is available

This function:
1. Creates mount point and mounts the Rosetta virtiofs share
2. Disables QEMU x86_64 binfmt registration (if present)
3. Registers Rosetta as the x86_64 binfmt handler
4. Restarts systemd-binfmt to pick up the new handler

### Layer 6: Preflight Check

**Files:**
- `pkg/crc/preflight/preflight_darwin.go` — Add a preflight check that Rosetta is installed on the host
- `pkg/crc/preflight/preflight_checks_darwin.go` — Implement check/fix functions

### Layer 7: Tests

**Files:**
- `pkg/crc/config/config_test.go` or new test file — Test that `use-rosetta` setting validates correctly
- `pkg/drivers/vfkit/driver_darwin_test.go` — Test that Rosetta device is added when `Rosetta = true`

## Summary of Changes by File

| File | Change |
|------|--------|
| `pkg/crc/config/settings.go` | Add `UseRosetta` constant + `AddSetting()` call |
| `pkg/crc/config/validations.go` | Add `validateRosetta()` (arm64+darwin guard) |
| `pkg/crc/machine/types/types.go` | Add `EnableRosetta bool` to `StartConfig` |
| `pkg/crc/machine/config/config.go` | Add `EnableRosetta bool` to `MachineConfig` |
| `cmd/crc/cmd/start.go` | Wire `use-rosetta` config into `StartConfig` |
| `pkg/crc/machine/start.go` | Pass to `MachineConfig` + add `configureRosetta()` + call it post-boot |
| `pkg/drivers/vfkit/driver_darwin.go` | Add `Rosetta bool` field + add `RosettaShare` device in `Start()` |
| `pkg/crc/machine/vfkit/driver_darwin.go` | Set `vfDriver.Rosetta` in `CreateHost()` |
| `pkg/crc/preflight/preflight_darwin.go` | Add Rosetta preflight check (conditional on config) |

## Key Design Decisions

1. **`RequiresDeleteAndSetupMsg`** — Changing `use-rosetta` requires deleting and recreating the VM because the vfkit Rosetta device must be present at VM launch. This matches how `preset` works.

2. **The vendored vfkit library already supports Rosetta** — `RosettaShareNew()` at `vendor/github.com/crc-org/vfkit/pkg/config/virtio.go:759` generates the correct `--device rosetta,mountTag=...` CLI args. No vendor changes needed.

3. **In-VM binfmt_misc setup happens every start** — The `/proc/sys/fs/binfmt_misc` registrations don't persist across VM reboots, so `configureRosetta()` must run on every `crc start`, not just on initial creation. This follows the same pattern as `configureSharedDirs()`.

4. **Platform gating** — The setting only validates on `darwin`+`arm64`. On Intel Macs or Linux, setting `use-rosetta true` returns an error.

## Risks / Open Questions

1. **SELinux context** — The mount command uses `context="system_u:object_r:container_file_t:s0"` matching the existing shared dirs pattern. This may need adjustment depending on the CoreOS/RHCOS SELinux policy in the bundle.

2. **Rosetta limitations** — As noted in the issue by @cfergeau, Rosetta has its own compatibility issues (no AVX/AVX2, some syscall gaps). The help text should mention this is a trade-off vs. QEMU.

3. **Persistence of binfmt** — If the user also wants QEMU for other architectures (e.g., arm32), disabling `qemu-x86_64` specifically (not all of binfmt) is the right approach. The implementation above is targeted.

4. **vfkit version** — The current vendored vfkit is `0.6.1`. Rosetta support in the Virtualization.framework requires macOS 13+. The existing `supportsVirtiofs()` check (macOS 12+) is close but a separate macOS 13+ check may be warranted.
