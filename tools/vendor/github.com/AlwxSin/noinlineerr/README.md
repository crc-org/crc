# noinlineerr

A go linter that forbids inline error handling using `if err := ...; err != nil`.

---

## Why?
Inline error handling in Go can hurt readibility by hiding the actual function call behind error plubming.
We believe errors and function deserve their own spotlight.

Instead of:
```go
if err := doSomething(); err != nil {
    return err
}
```
Prefer more explicit and readable:
```go
err := doSomething()
if err != nil {
    return err
}
```

---

## Install
```bash
go install github.com/AlwxSin/noinlineerr/cmd/noinlineerr@latest
```

---

## Usage
### As a standalone tool
```bash
noinlineerr ./...
```

⚠️ Note: the linter detects inline error assignments only when the error variable is explicitly typed or deducible. It doesn't handle dynamically typed interfaces (e.g. foo().Err() where Err() returns error via interface).

---

## Development
Run tests:
```bash
go test ./...
```
Test data lives under `testdata/src/...`

---

## Contributing
PRs welcome. Let's make Go code cleaner, one `err` at a time.
