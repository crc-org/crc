//go:build p9debug
// +build p9debug

package debug

import (
	"fmt"
	"os"
)

func Log(str string, args ...any) {
	fmt.Fprintf(os.Stderr, str, args...)
}
