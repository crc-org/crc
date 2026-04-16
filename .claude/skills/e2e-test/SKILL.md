---
name: e2e-test
description: Run CRC end-to-end tests for specific features and operating systems
argument-hint: [feature] [os]
allowed-tools: [AskUserQuestion, Bash, Read, Grep]
---

# CRC E2E Test Runner

This skill runs CRC end-to-end tests with interactive feature and OS selection.

## Arguments

User provided arguments: $ARGUMENTS

## Available Features

The CRC e2e test suite includes these features:
- **basic**: Core CRC lifecycle tests (version, help, setup, start, stop, delete)
- **config**: Configuration management tests
- **minimal**: Minimal feature set tests (requires MicroShift bundle)
- **story_openshift**: OpenShift-specific story tests
- **story_microshift**: MicroShift-specific story tests
- **story_application_deployment**: Application deployment scenarios
- **running_cluster_tests**: Tests for running cluster operations
- **cert_rotation**: Certificate rotation tests (Linux only)
- **story_manpages**: Man page generation tests

## Available OS Platforms

- **linux**: Linux platform tests
- **darwin**: macOS platform tests
- **windows**: Windows platform tests

## Instructions

When this skill is invoked:

1. **Parse arguments** (if provided):
   - If arguments include a feature name and/or OS, use those
   - Otherwise, proceed to interactive selection

2. **Interactive selection** (if no arguments or partial arguments):
   
   Use AskUserQuestion tool with these specific questions:
   
   **Question 1: Feature Selection**
   - Header: "Feature"
   - Question: "Which e2e test feature(s) do you want to run?"
   - multiSelect: true (allow multiple features)
   - Options:
     * basic: "Core CRC lifecycle tests - setup, start, stop, delete, version, status (~30 min)"
     * config: "Configuration management and property validation tests (~15 min)"
     * minimal: "Quick minimal test suite for fast validation (~10 min, requires MicroShift bundle)"
     * story_openshift: "OpenShift-specific features and scenarios (~45 min)"
     * story_microshift: "MicroShift-specific features and scenarios (~45 min)"
     * story_application_deployment: "Application deployment workflows (~30 min)"
     * running_cluster_tests: "Tests for running cluster operations (~20 min, requires running cluster)"
     * cert_rotation: "Certificate rotation scenarios (~25 min, Linux only)"
     * story_manpages: "Man page generation and validation (~10 min)"
   
   **Question 2: OS Platform Selection**
   - Header: "Platform"
   - Question: "Which OS platform do you want to test?"
   - multiSelect: false (single selection)
   - Options:
     * current: "Use current platform ($(GOOS)) - Recommended"
     * linux: "Linux platform (most features supported)"
     * darwin: "macOS platform (arm64 and amd64)"
     * windows: "Windows platform (amd64 only)"
   
   **Question 3: Bundle Location**
   - Header: "Bundle"
   - Question: "Where is the CRC bundle located?"
   - multiSelect: false (single selection)
   - Options:
     * cache: "~/.crc/cache/crc_*.crcbundle (CRC cache directory) - Recommended"
     * downloads: "~/Downloads/crc_*.crcbundle"
     * custom: "Specify custom path"
   
   **Question 4: Pull Secret Location**
   - Header: "Pull Secret"
   - Question: "Where is the pull secret file located?"
   - multiSelect: false (single selection)
   - Options:
     * downloads: "~/Downloads/crc-pull-secret (default) - Recommended"
     * custom: "Specify custom path"

3. **Handle custom paths** (if user selected "custom" for any location):
   - For each "custom" selection, the user will provide the path via "Other" option
   - Extract the custom path from the user's response
   - Validate that custom paths exist before proceeding
   - Store paths for use in environment variables

4. **Build CRC from source**:
   - Run `make cross` to build CRC binaries for all platforms
   - This will create platform-specific binaries in the `out/` directory:
     * Linux AMD64: `out/linux-amd64/crc`
     * Linux ARM64: `out/linux-arm64/crc`
     * macOS AMD64: `out/macos-amd64/crc`
     * macOS ARM64: `out/macos-arm64/crc`
     * Windows AMD64: `out/windows-amd64/crc.exe`
   - Display build progress to user
   - Verify the binary was built successfully for the target platform

5. **Determine CRC binary directory**:
   - Based on the selected platform, set the binary directory:
     * If platform is "current": detect using `go env GOOS` and `go env GOARCH`
     * If platform is "linux": use `--crc-binary=out/linux-amd64/`
     * If platform is "darwin": detect arch with `uname -m` (arm64 or amd64) and use `--crc-binary=out/macos-<arch>/`
     * If platform is "windows": use `--crc-binary=out/windows-amd64/`
   - Note: CRC_BINARY should be the directory path, not including the binary name itself
   - Verify the binary exists in the determined directory (e.g., `out/linux-amd64/crc` or `out/windows-amd64/crc.exe`)

6. **Prepare test environment**:
   - Based on bundle location selection:
     * If "cache": list available bundles in ~/.crc/cache and let user select or provide path
     * If "downloads": list available bundles in ~/Downloads and let user select or provide path
     * If "custom": ask user for the path
   - Verify the bundle file exists at the specified location
   - Verify the pull secret file exists at the specified location
   - Inform user about any missing prerequisites

7. **Clean CRC state** (IMPORTANT - Required for most tests):
   - Run `crc cleanup` using the built binary (e.g., `out/linux-amd64/crc cleanup`)
   - This ensures tests start with a clean state (no existing VM, no stale configuration)
   - **Critical for @basic tests** which expect an unconfigured system
   - Note: Cleanup may require sudo password for removing system files:
     * `/etc/udev/rules.d/99-crc-vsock.rules` (udev rules)
     * `/etc/modules-load.d/vhost_vsock.conf` (vsock module config)
   - If cleanup fails due to missing sudo access at the end of test run, it's acceptable (cleanup failure in tests is cosmetic)
   - Use absolute paths (not tilde) for bundle and pull secret locations to avoid path expansion issues

8. **Build test command**:
   
   Construct the make command as follows:
   
   a. Start with base: `make e2e`
   
   b. Build tag filter:
   - If platform is "current": use `$(GOOS)` or detect with `go env GOOS`
   - If single feature: `--godog.tags="<os> && @<feature>"`
   - If multiple features: `--godog.tags="<os> && (@feature1 || @feature2 || @feature3)"`
   - Always include OS tag AND feature tag(s)
   
   c. Build environment variables based on user selections:
   - Always set `CRC_BINARY` to the platform-specific binary directory from step 5 (format: `--crc-binary=/absolute/path/to/out/<platform>-<arch>/`)
   - For bundle location:
     * IMPORTANT: Use absolute paths (not tilde ~) to avoid path expansion issues
     * If "cache": prompt user to select specific bundle from ~/.crc/cache or set `BUNDLE_LOCATION=--bundle-location=/absolute/path/to/.crc/cache/<bundle_file>`
     * If "downloads": prompt user to select specific bundle from ~/Downloads or set `BUNDLE_LOCATION=--bundle-location=/absolute/path/to/Downloads/<bundle_file>`
     * If "custom": set `BUNDLE_LOCATION=--bundle-location=<absolute_custom_path>`
   - For pull secret location:
     * IMPORTANT: Use absolute paths (not tilde ~)
     * If "downloads": set `PULL_SECRET_FILE=--pull-secret-file=/absolute/path/to/Downloads/crc-pull-secret`
     * If "custom": set `PULL_SECRET_FILE=--pull-secret-file=<absolute_custom_path>`
   - Note: All environment variables must include their flag prefix and use absolute paths
   
   d. Combine into final command:
   ```bash
   CRC_BINARY=--crc-binary=<binary_dir> [OTHER_ENV_VARS] make e2e GODOG_OPTS="--godog.tags=\"<constructed_tags>\""
   ```
   
   Note: `<binary_dir>` is the absolute directory path (e.g., `/home/user/crc/out/linux-amd64/`), not the full binary path
   
   e. Example commands:
   ```bash
   # Single feature on Linux with built binary and bundle from cache (using absolute paths)
   CRC_BINARY=--crc-binary=/home/user/crc/out/linux-amd64/ BUNDLE_LOCATION=--bundle-location=/home/user/.crc/cache/crc_libvirt_4.21.8_amd64.crcbundle PULL_SECRET_FILE=--pull-secret-file=/home/user/pull-secret make e2e GODOG_OPTS="--godog.tags=\"linux && @basic\""
   
   # Multiple features on macOS with built binary
   CRC_BINARY=--crc-binary=/Users/user/crc/out/macos-arm64/ BUNDLE_LOCATION=--bundle-location=/Users/user/.crc/cache/crc_vfkit_4.21.8_arm64.crcbundle make e2e GODOG_OPTS="--godog.tags=\"darwin && (@basic || @config || @minimal)\""
   ```

9. **Execute tests**:
   - Run the test command
   - Display progress and results to user
   - Provide clear feedback on pass/fail status
   - Note: If the final test cleanup step fails due to sudo password requirement, this is acceptable
     * The test suite runs its own cleanup at the end which may fail in non-interactive environments
     * Focus on the actual test results (scenarios/steps passed), not cleanup failure
     * A test is considered successful if all scenarios passed, even if final cleanup failed

## Environment Variables

These are set automatically by the skill:
- `CRC_BINARY`: Directory path to platform-specific CRC binary built from source (e.g., `--crc-binary=out/linux-amd64/`)
- `PULL_SECRET_FILE`: Path to pull secret (e.g., `--pull-secret-file=~/pull-secret` or default if not specified)
- `BUNDLE_LOCATION`: Path to CRC bundle (e.g., `--bundle-location=~/.crc/cache/crc_*.crcbundle`)
- `CLEANUP_HOME`: Whether to cleanup home directory (default: not set)

Note: All environment variables include the flag prefix (e.g., `--crc-binary=`, `--bundle-location=`, `--pull-secret-file=`)

## Example Usage

```bash
# Interactive mode
/e2e-test

# With feature argument
/e2e-test basic

# With feature and OS
/e2e-test config linux

# Multiple features
/e2e-test basic,config darwin
```

## Important Notes

### Test Duration
- Tests can take 10-180 minutes depending on feature selection
- Full test suite (@basic) typically takes ~30 minutes
- Multiple features will run sequentially

### Prerequisites
- **Build environment**: Go toolchain and build dependencies for `make cross`
- **Pull secret**: Required at ~/Downloads/crc-pull-secret (or custom path)
- **CRC bundle**: Required at ~/.crc/cache/crc_*.crcbundle, ~/Downloads/crc_*.crcbundle, or custom path
- **Clean state**: Most tests expect no existing CRC cluster
- **System resources**: Ensure adequate RAM (16GB+ recommended for monitoring tests)
- **Build time**: `make cross` takes ~5-10 minutes to compile binaries for all platforms

### Platform-Specific Considerations
- `@cert_rotation` only runs on Linux
- `@running_cluster_tests` requires a cluster to already be running
- Windows tests require `.exe` extension handling
- macOS supports both arm64 and amd64 architectures
- `@minimal` test requires MicroShift bundle (crc_microshift_*.crcbundle), not OpenShift bundle

### Key Technical Notes
- **Path Expansion**: Always use absolute paths (e.g., `/home/user/.crc/cache/...`) instead of tilde paths (`~/.crc/cache/...`) to avoid shell expansion issues in the test framework
- **Cleanup Requirements**: Running `crc cleanup` before tests ensures clean state, especially critical for `@basic` tests that check for unconfigured system
- **Sudo Access**: Cleanup operations require sudo to remove system files (udev rules, vsock config); have password ready
- **Test Cleanup Failures**: If the test suite's final cleanup step fails due to sudo, it's cosmetic - focus on scenario/step pass rates

### Before Running
The skill will automatically:
1. Build CRC from source using `make cross`
2. Verify the binary exists for the target platform
3. Check that bundle and pull secret files exist
4. **Run `crc cleanup`** to ensure clean state (may require sudo password)

The user should ensure:
1. **Sudo access available** (cleanup requires sudo to remove system files like udev rules)
2. Pull secret and bundle files are available (check ~/.crc/cache or ~/Downloads)
3. Adequate disk space for build artifacts (~500MB in `out/` directory)
4. **No active CRC cluster** is running (will be cleaned up automatically)

### Execution Behavior
- **Build phase**: `make cross` compiles binaries for all platforms (~5-10 minutes)
- **Cleanup phase**: `crc cleanup` removes existing CRC state (may prompt for sudo password)
- **Test phase**: Tests run with `--timeout=180m` (3 hours max)
- Output is verbose (`-v` flag enabled)
- Tests may modify CRC configuration
- Some tests require internet connectivity
- Failed tests will show detailed error output
- The skill always uses the freshly built binary from the `out/` directory
- **IMPORTANT**: Always use absolute paths (not `~`) for bundle and pull secret locations to avoid path expansion issues
