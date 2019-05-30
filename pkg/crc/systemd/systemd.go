package systemd

import (
	"github.com/code-ready/crc/pkg/crc/systemd/states"
)

type SystemdCommander interface {
	Start(name string) (bool, error)
	Stop(name string) (bool, error)
	Reload(name string) (bool, error)
	Restart(name string) (bool, error)
	Status(name string) (states.State, error)
	Enable(name string) (bool, error)
	Disable(name string) (bool, error)
	DaemonReload() (bool, error)
}
