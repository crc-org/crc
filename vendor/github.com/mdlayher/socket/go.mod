module github.com/mdlayher/socket

go 1.17

require (
	github.com/google/go-cmp v0.5.7
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8
)

// Intentional pin to an older release; we don't need new features and would
// like to support older verions of Go.
require golang.org/x/net v0.0.0-20190503192946-f4e77d36d62c
