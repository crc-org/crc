# AGENTS.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

CRC (Runs Containers) manages a local OpenShift 4.x, OKD, or MicroShift cluster VM for development and testing. Module path: `github.com/crc-org/crc/v2`, Go 1.25.

## Build Commands

```bash
make install              # Build and install crc binary
make cross                # Cross-compile for all platforms (linux/darwin/windows, amd64/arm64)
make test                 # Unit tests: go test -race --tags "build containers_image_openpgp" ./pkg/... ./cmd/...
make lint                 # Run golangci-lint
make cross-lint           # Lint for all target platforms
make fmt                  # Format with goimports
make vendor               # go mod tidy + go mod vendor
make vendorcheck          # Verify vendor directory is up to date
make generate_mocks       # Generate mocks with mockery
make check                # Full CI: cross build + tests + lint + vendorcheck
```

### Running a single test

```bash
go test -race --tags "build containers_image_openpgp" -run TestFunctionName ./pkg/crc/machine/...
```

Build tags `build` and `containers_image_openpgp` are required. Version variables are injected via `-ldflags` at build time (see `VERSION_VARIABLES` in Makefile).

### E2E and integration tests

```bash
make e2e GODOG_OPTS="--godog.tags=\"linux && @basic\""     # BDD tests (Godog/Cucumber)
make integration                                            # Ginkgo-based integration tests
```

Both require a pull secret and CRC bundle. E2E features are in `test/e2e/features/*.feature`, tagged by platform (`@linux`, `@darwin`, `@windows`) and feature (`@basic`, `@config`, `@minimal`, etc.).

## Architecture

### Entry points

- **`cmd/crc/`** — Main CLI binary (Cobra). `cmd/crc/cmd/` has subcommands: `start`, `stop`, `delete`, `status`, `setup`, `config`, `bundle`, `daemon`.
- **`cmd/crc-embedder/`** — Utility for embedding helper binaries into CRC during release builds.

### Core packages (pkg/crc/)

- **`machine/`** — Central package. Defines `Client` interface for VM lifecycle (Start/Stop/Delete/Status). Platform-specific drivers selected via `driver_linux.go`, `driver_darwin.go`, `driver_windows.go`.
- **`config/`** — Viper-based configuration storage, validation, and defaults.
- **`cluster/`** — Kubernetes/OpenShift operations: pull secret handling, cert renewal, CSR approval, cluster operator status, kubeadmin password management.
- **`api/`** — HTTP REST API server that the daemon exposes. `api/client/` provides the client library. `api/events/` implements SSE streaming.
- **`preflight/`** — Platform-specific system requirement checks (hypervisor, CPU, memory, network). Runs before cluster start.
- **`preset/`** — Multi-preset support: OpenShift, OKD, MicroShift.
- **`daemonclient/`** — Client for CLI→daemon communication over Unix socket (NetworkClient + APIClient + SSEClient).
- **`cache/`** — Bundle download and extraction caching, platform-specific implementations.

### Hypervisor drivers

- `pkg/crc/machine/libvirt/` — Linux (libvirt/KVM)
- `pkg/crc/machine/libhvee/` — Windows (Hyper-V)
- `pkg/drivers/vfkit/` — macOS (Virtualization.framework)

### Request flow

CLI command → `daemonclient` → HTTP over Unix socket → daemon `api.Handler` → `machine.Client` → platform driver → VM

### Platform-specific code

Files with `_linux.go`, `_darwin.go`, `_windows.go` suffixes throughout `cmd/crc/cmd/`, `pkg/crc/machine/`, `pkg/crc/preflight/`, `pkg/crc/constants/`, and `pkg/crc/cache/`.

## Conventions

- Commit message format: `area: Short description` (e.g., `daemon: Remove tcp fallback for 9p file sharing`)
- Tools are managed via `tool` directives in `go.mod` — run tools with `go tool <name>` (golangci-lint, goimports, mockery, gomod2rpmdeps, makefat)
- Linter config: `.golangci.yml` (golangci-lint v2, 10min timeout, build tags: `build` + `containers_image_openpgp`)
- Mocks generated with mockery into `test/mocks/`
