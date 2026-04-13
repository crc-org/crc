# E2E Test Skill Examples

## Interactive Mode (Recommended)

Simply invoke the skill without arguments:

```bash
/e2e-test
```

The skill will:
1. Ask you to select features (multi-select enabled)
2. Ask you to select OS platform
3. Ask for bundle location (~/.crc/cache, ~/Downloads, or custom)
4. Ask for pull secret location (~/Downloads or custom)
5. Build CRC from source using `make cross` (~5-10 minutes)
6. Verify prerequisites
7. Build and execute the test command with the built binary
8. Show results

### Example Interactive Flow

```
User: /e2e-test

Claude: Which e2e test feature(s) do you want to run?
[✓] basic - Core CRC lifecycle tests
[ ] config - Configuration management tests
[✓] minimal - Quick minimal test suite
[ ] ...

Claude: Which OS platform do you want to test?
( ) linux
(•) darwin (current)
( ) windows

Claude: Running: make e2e GODOG_OPTS="--godog.tags=\"darwin && (@basic || @minimal)\""
...
```

## With Arguments

### Single Feature, Current OS

```bash
/e2e-test basic
```

Runs basic tests on current platform.

### Single Feature, Specific OS

```bash
/e2e-test config linux
```

Runs config tests on Linux platform.

### Multiple Features

```bash
/e2e-test basic,minimal,config
```

Prompts for OS selection, then runs all three features.

### Multiple Features with OS

```bash
/e2e-test basic,config darwin
```

Runs basic and config tests on macOS.

## Common Scenarios

### Quick Validation

Run minimal tests on current platform:
```bash
/e2e-test minimal
```
~10 minutes

### Core Feature Testing

Run basic lifecycle tests:
```bash
/e2e-test basic linux
```
~30 minutes

### OpenShift Story Validation

```bash
/e2e-test story_openshift
```
~45 minutes, tests OpenShift-specific features

### Configuration Testing

```bash
/e2e-test config
```
~15 minutes, validates configuration management

### Multi-Feature Test Run

```bash
/e2e-test basic,config,minimal linux
```
~55 minutes combined

## Advanced Usage

### Testing on Running Cluster

For tests that require an existing running cluster:

```bash
# First, ensure cluster is running
crc start

# Then run tests that don't manage lifecycle
/e2e-test running_cluster_tests
```

### Cleanup After Tests

```bash
# After running tests, cleanup
crc delete -f
crc cleanup
```

## Expected Output

### Successful Test Run

```
Building CRC from source...
Running: make cross

Building binaries for all platforms...
✓ out/linux-amd64/crc
✓ out/linux-arm64/crc
✓ out/macos-amd64/crc
✓ out/macos-arm64/crc
✓ out/windows-amd64/crc.exe

Build completed in 6m32s

Checking prerequisites...
✓ Pull secret found at ~/pull-secret
✓ Bundle found at ~/.crc/cache/crc_libvirt_4.21.8_amd64.crcbundle
✓ CRC binary built at out/linux-amd64/crc

Running e2e tests:
  Features: basic, config
  Platform: linux
  
Command: CRC_BINARY=--crc-binary=out/linux-amd64/ BUNDLE_LOCATION=--bundle-location=~/.crc/cache/crc_libvirt_4.21.8_amd64.crcbundle PULL_SECRET_FILE=--pull-secret-file=~/pull-secret make e2e GODOG_OPTS="--godog.tags=\"linux && (@basic || @config)\""

=== Test Output ===
go test --timeout=180m github.com/crc-org/crc/v2/test/e2e -tags "containers_image_openpgp" ...

Feature: Basic test
  Scenario: CRC version                    # test/e2e/features/basic.feature:8
    ✓ crc version has expected output

  Scenario: CRC help                       # test/e2e/features/basic.feature:12
    ✓ executing crc help command succeeds
    ✓ stdout should contain "Usage:"
    ...

Feature: Config test
  Scenario: Get config value              # test/e2e/features/config.feature:10
    ✓ executing "crc config get memory" succeeds
    ...

32 scenarios (32 passed)
156 steps (156 passed)
28m15.234s

Tests completed successfully!
```

### Test Failure

```
...
  Scenario: CRC start usecase              # test/e2e/features/basic.feature:31
    ✓ executing "crc setup --check-only" fails
    ✓ unsetting config property "developer-password" succeeds
    ✓ setting config property "enable-cluster-monitoring" to value "true" succeeds
    ✗ starting CRC with default bundle succeeds
      Error: timeout waiting for cluster to start

--- Failed scenarios:
    test/e2e/features/basic.feature:31

3 scenarios (2 passed, 1 failed)
12 steps (11 passed, 1 failed)
```

## Troubleshooting

### "Bundle not found"

```bash
# Check if bundle exists in CRC cache directory
ls ~/.crc/cache/*.crcbundle

# Or check Downloads folder
ls ~/Downloads/*.crcbundle

# Or download bundle and it will be stored in cache
crc setup

# Or specify custom path when prompted by the skill
```

### "Pull secret not found"

```bash
# Get pull secret from https://console.redhat.com/openshift/create/local
# Save to ~/Downloads/crc-pull-secret or specify custom path when prompted
# Example:
# mv ~/Downloads/pull-secret.txt ~/Downloads/crc-pull-secret
```

### "CRC is already running"

```bash
# Stop and delete existing cluster
crc stop
crc delete -f
```

### Tests timeout

- Increase system resources
- Check network connectivity
- Verify bundle is not corrupted
- Try running fewer features at once

## Tips

1. **Start small**: Test with `minimal` or single features first
2. **Build time**: First run takes ~5-10 minutes for `make cross`, subsequent runs reuse built binaries if source hasn't changed
3. **Bundle location**: Use ~/.crc/cache for bundles (recommended) - this is where `crc setup` downloads them
4. **Clean state**: Most tests expect no running cluster - use `crc delete -f` before running
5. **Monitor resources**: Watch system RAM and disk usage during tests (16GB+ RAM recommended)
6. **Logs**: Check `~/.crc/crc.log` for detailed CRC logs if tests fail
7. **Parallel testing**: Don't run multiple e2e test sessions simultaneously
8. **Fresh build**: The skill always builds from current source, ensuring tests run against your latest changes
