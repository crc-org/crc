package input

import (
	"fmt"
	"strings"

	"github.com/code-ready/crc/pkg/crc/output"
	survey "gopkg.in/AlecAivazis/survey.v1"
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

// PromptUserForSecret can be used for any kind of secret like image pull
// secret or for password.
func PromptUserForSecret(message string, help string) (string, error) {
	var secret string
	prompt := &survey.Password{
		Message: message,
		Help:    help,
	}
	if err := survey.AskOne(prompt, &secret, nil); err != nil {
		return "", err
	}
	return secret, nil
}
