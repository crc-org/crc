// This file is not meant to be built, it's just a placeholder to ensure
// the source for the various tools that we use is properly referenced in go.mod
// and vendored in vendor/
package buildtools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/randall77/makefat"
)
