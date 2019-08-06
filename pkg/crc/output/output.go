package output

import (
	"fmt"
	"io"
)

func Out(args ...interface{}) {
	fmt.Println(args...)
}

func OutF(s string, args ...interface{}) {
	fmt.Printf(s, args...)
}

func OutW(w io.Writer, args ...interface{}) (n int, err error) {
	return fmt.Fprintln(w, args...)
}
