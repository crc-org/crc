# E2E Test Reference

## Feature Tags and Descriptions

Based on the test/e2e/features directory:

### Core Features

| Feature Tag | File | Description | Platform Support |
|------------|------|-------------|------------------|
| `@basic` | basic.feature | Core CRC lifecycle (setup, start, stop, delete, version, help, status) | linux, darwin, windows |
| `@config` | config.feature | Configuration property management and validation | linux, darwin, windows |
| `@minimal` | minimal.feature | Minimal test suite for quick validation | linux, darwin, windows |

### Story Features

| Feature Tag | File | Description | Platform Support |
|------------|------|-------------|------------------|
| `@story_openshift` | story_openshift.feature | OpenShift-specific features and scenarios | linux, darwin, windows |
| `@story_microshift` | story_microshift.feature | MicroShift-specific features and scenarios | linux, darwin, windows |
| `@story_application_deployment` | application_deployment.feature | Application deployment workflows | linux, darwin, windows |
| `@story_manpages` | manpages.feature | Man page generation and validation | linux, darwin, windows |

### Specialized Tests

| Feature Tag | File | Description | Platform Support |
|------------|------|-------------|------------------|
| `@running_cluster_tests` | running_cluster_tests.feature | Tests for operations on running clusters | linux, darwin, windows |
| `@cert_rotation` | cert_rotation.feature | Certificate rotation scenarios | linux only |

## Additional Tags

### Platform Tags
- `@darwin` - macOS tests
- `@linux` - Linux tests  
- `@windows` - Windows tests

### Lifecycle Tags
- `@startstop` - Tests that manage cluster start/stop
- `@cleanup` - Tests that require cleanup
- `@release` - Tests included in release validation

## Makefile Targets

### Main E2E Target
```bash
make e2e
```

**Default behavior:**
- Runs with OS tag matching current platform (GOOS)
- Uses pull secret from ~/Downloads/crc-pull-secret
- Uses bundle from ~/Downloads/crc_libvirt_*.crcbundle (or ~/.crc/cache)
- Uses CRC binary from $GOPATH/bin

**E2E Test Skill behavior:**
- Builds CRC from source using `make cross` before running tests
- Uses built binary from `out/<platform>-<arch>/crc`
- Prompts for bundle location (cache, downloads, or custom)
- Prompts for pull secret location (downloads or custom)

### Story-Specific Targets
These run with `--cleanup-home=false` (cluster must already be running):

```bash
make e2e-story-health
make e2e-story-marketplace  
make e2e-story-registry
make e2e-story-microshift
```

### Tag Filtering Examples

Run specific feature on current OS:
```bash
make e2e GODOG_OPTS="--godog.tags=\"\$(GOOS) && @basic\""
```

Run multiple features:
```bash
make e2e GODOG_OPTS="--godog.tags=\"linux && (@basic || @config)\""
```

Exclude tags:
```bash
make e2e GODOG_OPTS="--godog.tags=\"linux && ~@startstop\""
```

## Environment Overrides

```bash
# Custom pull secret
make e2e PULL_SECRET_FILE=--pull-secret-file=/custom/path/secret

# Bundle from CRC cache directory (recommended)
make e2e BUNDLE_LOCATION=--bundle-location=~/.crc/cache/crc_libvirt_4.21.8_amd64.crcbundle

# Bundle from Downloads
make e2e BUNDLE_LOCATION=--bundle-location=~/Downloads/crc_libvirt_4.21.8_amd64.crcbundle

# Custom CRC binary from build output (directory path, not binary itself)
make e2e CRC_BINARY=--crc-binary=out/linux-amd64/

# Skip home cleanup
make e2e CLEANUP_HOME=--cleanup-home=false

# Test specific version
make e2e VERSION_TO_TEST=--test-version=v2.58.0

# E2E Test Skill automatically sets these (note: directory path for CRC_BINARY):
CRC_BINARY=--crc-binary=out/linux-amd64/ BUNDLE_LOCATION=--bundle-location=~/.crc/cache/crc_libvirt_4.21.8_amd64.crcbundle PULL_SECRET_FILE=--pull-secret-file=~/pull-secret make e2e GODOG_OPTS="--godog.tags=\"linux && @minimal\""
```

## Test Execution Notes

- **Timeout**: Tests use `--timeout=180m` (3 hours)
- **Build tags**: Tests are built with container_image_openpgp tag
- **Prerequisites**: CRC must be properly setup with required resources
- **Clean state**: Most tests expect a clean starting state (no existing cluster)
