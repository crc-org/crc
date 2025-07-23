package real

import (
	"fmt"
	"os"
)

func getUserGroup(info os.FileInfo) (string, string, error) {
	return "", "", fmt.Errorf("Cannot get system-specific info on Windows.")
}
