package goodhosts

import (
	"strings"

	"github.com/code-ready/crc/pkg/os/windows/powershell"
)

func execute(args ...string) error {
	_, _, err := powershell.ExecuteAsAdmin("modifying hosts file", strings.Join(append([]string{goodhostPath}, args...), " "))
	return err
}
