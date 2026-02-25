package terminal

import (
	"math"
	"os"

	xterminal "golang.org/x/term"
)

var (
	// Global variable to force output regardless if terninal
	ForceShowOutput = false
)

func IsShowTerminalOutput() bool {
	// if this is a terminal or set to force output
	return IsRunningInTerminal() || ForceShowOutput
}

func IsRunningInTerminal() bool {
	fd := os.Stdout.Fd()
	if fd > math.MaxInt {
		return false
	} else {
		return xterminal.IsTerminal(int(fd))
	}
}
