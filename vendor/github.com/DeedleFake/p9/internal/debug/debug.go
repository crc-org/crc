// +build p9debug

package debug

import (
	"fmt"
	"os"
)

func Log(str string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, str, args...)
}
