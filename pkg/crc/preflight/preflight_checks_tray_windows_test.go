package preflight

import (
	"fmt"
	"strings"
	"testing"

	"github.com/code-ready/crc/pkg/os/windows/powershell"
)

func TestEscapeWindowsPassword(t *testing.T) {
	passwords := []string{
		"tes;t@\"_\\/",
		"$kdhhjs;%&*'`",
		"``````$''\"",
	}

	for _, pass := range passwords {
		op := escapeWindowsPassword(pass)
		psCmd := fmt.Sprintf("$var=\"%s\"; $var", op)
		stdOut, stdErr, err := powershell.Execute(psCmd)
		if err != nil {
			t.Errorf("Error while executing powershell command: %v: %v", err, stdErr)
		}
		if strings.TrimSpace(stdOut) != pass {
			t.Errorf("Passwords don't match after escaping for powershell. Expected: %s, Actual: %s", pass, stdOut)
		}
	}
}
