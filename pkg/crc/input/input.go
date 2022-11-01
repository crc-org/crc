package input

import (
	"fmt"
	"strings"

	crcos "github.com/crc-org/crc/pkg/os"
)

func PromptUserForYesOrNo(message string, force bool) bool {
	if force {
		return true
	}
	if !crcos.RunningInTerminal() {
		return false
	}
	var response string
	fmt.Printf(message + "? [y/N]: ")
	fmt.Scanf("%s", &response)

	return strings.ToLower(response) == "y"
}
