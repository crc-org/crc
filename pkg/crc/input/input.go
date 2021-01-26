package input

import (
	"fmt"
	"os"
	"strings"

	terminal "golang.org/x/term"
)

func PromptUserForYesOrNo(message string, force bool) bool {
	if force {
		return true
	}
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}
	var response string
	fmt.Printf(message + "? [y/N]: ")
	fmt.Scanf("%s", &response)

	return strings.ToLower(response) == "y"
}
