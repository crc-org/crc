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
https://github.com/sonatard/noctx/blob/03bbcad02284bb6257428c0e5d489e0d113bfee8/noctx.go#L41-L50

### Sample
https://github.com/sonatard/noctx/blob/03bbcad02284bb6257428c0e5d489e0d113bfee8/testdata/src/http_client/http_client.go#L11
https://github.com/sonatard/noctx/blob/03bbcad02284bb6257428c0e5d489e0d113bfee8/testdata/src/http_request/http_request.go#L17

### Reference
- [net/http - NewRequest](https://pkg.go.dev/net/http#NewRequest)
- [net/http - NewRequestWithContext](https://pkg.go.dev/net/http#NewRequestWithContext)
- [net/http - Request.WithContext](https://pkg.go.dev/net/http#Request.WithContext)

## net package

### Rules
https://github.com/sonatard/noctx/blob/03bbcad02284bb6257428c0e5d489e0d113bfee8/noctx.go#L26-L39

### Sample
https://github.com/sonatard/noctx/blob/03bbcad02284bb6257428c0e5d489e0d113bfee8/testdata/src/network/net.go#L15

### References
- [net - ListenConfig](https://pkg.go.dev/net#ListenConfig)
- [net - Dialer.DialContext](https://pkg.go.dev/net#Dialer.DialContext)
- [net - Resolver](https://pkg.go.dev/net#Resolver)
- [net - DefaultResolver](https://pkg.go.dev/net#DefaultResolver)

## database/sql package
### Rules
https://github.com/sonatard/noctx/blob/03bbcad02284bb6257428c0e5d489e0d113bfee8/noctx.go#L52-L66

### Sample
https://github.com/sonatard/noctx/blob/03bbcad02284bb6257428c0e5d489e0d113bfee8/testdata/src/sql/sql.go#L13

### Reference
- [database/sql](https://pkg.go.dev/database/sql)

## crypt/tls package
### Rules
https://github.com/sonatard/noctx/blob/03bbcad02284bb6257428c0e5d489e0d113bfee8/noctx.go#L68-L71

### Sample
https://github.com/sonatard/noctx/blob/03bbcad02284bb6257428c0e5d489e0d113bfee8/testdata/src/crypto_tls/tls.go#L18

### Reference
- [crypto/tls - Dialer.DialContext](https://pkg.go.dev/crypto/tls#Dialer.DialContext)
- [crypto/tls - Conn.HandshakeContext](https://pkg.go.dev/crypto/tls#Conn.HandshakeContext)