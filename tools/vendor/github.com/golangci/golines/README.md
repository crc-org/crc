# Fork of golines

This is a fork of [golines](https://github.com/segmentio/golines) to be usable as a library.

I considered other options before deciding to fork, but there are no straightforward or non-invasive changes.

Simply isolating some code within a package will not address the root issue unless the code is converted to a dedicated module and all debug/graph-related code is removed.
Realistically, this is not possible.

Issues should be open either on the original [golines repository](https://github.com/segmentio/golines) or on [golangci-lint repository](https://github.com/golangci/golangci-lint).

**No modifications will be accepted other than the synchronization of the fork.**

The synchronization of the fork will be done by the golangci-lint maintainers only.

## Modifications

- The original code is under the `main` package -> uses `golines` package.
- Some files have been removed to reduce dependencies:
  - `main.go`, `main_test.go`
  - `graph.go`, `graph_generated.go`, `graph.test`
  - `diff.go`, `diff_test.go`
  - `doc.go`
- Some other files have been removed because unused by the fork:
  - `.goreleaser.yaml`, `.pre-commit-hooks.yaml`
  - `.github/build.yml`, `.github/release.yml`
  - `Makefile`
- The workflow files (`lint.yml` and `test.yml`) are modified to run for this fork.
- The file `shortener.go` has been modified:
  - The `baseFormatterCmd` is hardcoded.
  - The code related to debug logs has been removed.
  - The code related to graph has been removed.
- The module name has been changed to `github.com/golangci/golines` to avoid replacement directives inside golangci-lint.

## History

- sync with fc305205784a70b4cfc17397654f4c94e3153ce4 (after v0.12.2)
