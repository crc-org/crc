package logging

import (
	"github.com/code-ready/crc/pkg/crc/output"
)

func Log(s string, args ... interface{}) {
	// temporary solution until log framework has been decided
	output.Out(s, args)
}
