# noctx

![](https://github.com/sonatard/noctx/workflows/CI/badge.svg)

`noctx` finds function calls without context.Context.

If you are using net/http package and sql/database package, you should use noctx.
Passing `context.Context` enables library user to cancel request, getting trace information and so on.

## Usage

### noctx with go vet

go vet is a Go standard tool for analyzing source code.

1. Install noctx.
```sh
$ go install github.com/sonatard/noctx/cmd/noctx@latest
```

2. Execute noctx
```sh
$ go vet -vettool=`which noctx` main.go
./main.go:6:11: net/http.Get must not be called
```

### noctx with golangci-lint

golangci-lint is a fast Go linters runner.

1. Install golangci-lint.
[golangci-lint - Install](https://golangci-lint.run/usage/install/)

2. Setup .golangci.yml
```yaml:
# Add noctx to enable linters.
linters:
  enable:
    - noctx

# Or enable-all is true.
linters:
  default: all
  disable:
   - xxx # Add unused linter to disable linters.
```

3. Execute noctx
```sh
# Use .golangci.yml
$ golangci-lint run

# Only execute noctx
golangci-lint run --enable-only noctx
```

## net/http package
### Rules
https://github.com/sonatard/noctx/blob/e9e23da29379b87a39ce50fd1ef7b273fee2461a/noctx.go#L28-L36

### Sample
https://github.com/sonatard/noctx/blob/9a514098df3f8a88e0fd6949320c4e0aa51b520c/testdata/src/http_client/http_client.go#L11
https://github.com/sonatard/noctx/blob/9a514098df3f8a88e0fd6949320c4e0aa51b520c/testdata/src/http_request/http_request.go#L17

### Reference
- [net/http - NewRequest](https://pkg.go.dev/net/http#NewRequest)
- [net/http - NewRequestWithContext](https://pkg.go.dev/net/http#NewRequestWithContext)
- [net/http - Request.WithContext](https://pkg.go.dev/net/http#Request.WithContext)
 
## database/sql package
### Rules
https://github.com/sonatard/noctx/blob/a00128b6a4087639ed0d13a123d0f9960309824f/noctx.go#L40-L48

### Sample
https://github.com/sonatard/noctx/blob/6e0f6bb8de1bd8a3c6e73439614927fd59aa0a8a/testdata/src/sql/sql.go#L13

### Reference
- [database/sql](https://pkg.go.dev/database/sql)
