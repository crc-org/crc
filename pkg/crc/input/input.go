package input

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/code-ready/crc/pkg/crc/output"
	"golang.org/x/crypto/ssh/terminal"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

func PromptUserForYesOrNo(message string, force bool) bool {
	if force {
		return true
	}
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}
	var response string
	output.Outf(message + "? [y/N]: ")
	fmt.Scanf("%s", &response)

	return strings.ToLower(response) == "y"
}

// PromptUserForSecret can be used for any kind of secret like image pull
// secret or for password.
func PromptUserForSecret(message string, help string) (string, error) {
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		return "", errors.New("cannot ask for secret, crc not launched by a terminal")
	}

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
