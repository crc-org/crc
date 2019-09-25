package output

import (
	"fmt"
	"io"
)

func Outln(args ...interface{}) {
	fmt.Println(args...)
}

func Outf(s string, args ...interface{}) {
	fmt.Printf(s, args...)
}

func Fout(w io.Writer, args ...interface{}) (n int, err error) {
	return fmt.Fprintln(w, args...)
}
