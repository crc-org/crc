package exit

import (
	"fmt"
	"os"

	"github.com/code-ready/crc/pkg/crc/output"
)

// WithMessage prints the specified message and then exits the program with the specified exit code.
// If the exit code is 0, the message is prints to stdout, otherwise to stderr.
func WithMessage(code int, text string, args ...interface{}) {
	if code == 0 {
		_, _ = output.Fout(os.Stdout, fmt.Sprintf(text, args...))
	} else {
		_, _ = output.Fout(os.Stderr, fmt.Sprintf(text, args...))
	}
	os.Exit(code)
}
