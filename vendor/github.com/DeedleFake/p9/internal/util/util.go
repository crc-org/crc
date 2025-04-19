package util

import (
	"fmt"
	"io"
)

// LimitedReader is a reimplementation of io.LimitedReader with two
// main differences:
//
// * N is a uint32, allowing for larger sizes on 32-bit systems.
// * A custom error can be returned if N becomes zero.
type LimitedReader struct {
	R io.Reader
	N uint32
	E error
}

func (lr LimitedReader) err() error {
	if lr.E == nil {
		return io.EOF
	}

	return lr.E
}

func (lr *LimitedReader) Read(buf []byte) (int, error) {
	if lr.N <= 0 {
		return 0, lr.err()
	}

	if uint32(len(buf)) > lr.N {
		buf = buf[:lr.N]
	}

	n, err := lr.R.Read(buf)
	lr.N -= uint32(n)
	return n, err
}

// Errorf is a variant of fmt.Errorf that returns an error being
// wrapped directly if it is one of a number of specific values, such
// as nil or io.EOF.
func Errorf(str string, args ...interface{}) error {
	for _, arg := range args {
		if arg == io.EOF {
			return arg.(error)
		}
	}

	return fmt.Errorf(str, args...)
}
