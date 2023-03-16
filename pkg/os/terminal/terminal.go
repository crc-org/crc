package terminal

import (
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
	return xterminal.IsTerminal(int(os.Stdout.Fd()))
}
