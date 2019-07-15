package input

import (
	"fmt"
	"github.com/code-ready/crc/pkg/crc/output"
	"strings"
)

func PromptUserForYesOrNo(message string, force bool) bool {
	if force {
		return true
	}
	var response string
	output.OutF(message + "? [y/N]: ")
	fmt.Scanf("%s", &response)
	if strings.ToLower(response) == "y" {
		return true
	}
	return false
}
