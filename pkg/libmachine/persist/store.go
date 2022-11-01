package persist

import (
	"github.com/crc-org/crc/pkg/libmachine/host"
)

type Store interface {
	// SetExists defines whether a machine exists or not
	SetExists(name string) error

	// Exists returns whether a machine exists or not
	Exists(name string) (bool, error)

	// Load loads a host by name
	Load(name string) (*host.Host, error)

	// Remove removes a machine from the store
	Remove(name string) error

	// Save persists a machine in the store
	Save(host *host.Host) error
}
