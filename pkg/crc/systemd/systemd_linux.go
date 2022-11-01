package systemd

import (
	"os"
	"path/filepath"

	crcos "github.com/crc-org/crc/pkg/os"
)

func NewHostSystemdCommander() *Commander {
	return &Commander{
		commandRunner: crcos.NewLocalCommandRunner(),
	}
}

func UserUnitsDir() string {
	userConfigDir, _ := os.UserConfigDir()
	return filepath.Join(userConfigDir, "systemd", "user")
}

func UserUnitPath(unitName string) string {
	return filepath.Join(UserUnitsDir(), unitName)
}
