package systemd

import crcos "github.com/code-ready/crc/pkg/os"

func NewHostSystemdCommander() *Commander {
	return &Commander{
		commandRunner: crcos.NewLocalCommandRunner(),
	}
}
