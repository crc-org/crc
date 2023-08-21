package input

import (
	"fmt"
	"strings"

	crcTerminal "github.com/crc-org/crc/v2/pkg/os/terminal"
)

func PromptUserForYesOrNo(message string, force bool) bool {
	if force {
		return true
	}
	if !crcTerminal.IsRunningInTerminal() {
		return false
	}
	var response string
	fmt.Printf(message + "? [y/N]: ")
	fmt.Scanf("%s", &response)

	return strings.ToLower(response) == "y"
}
