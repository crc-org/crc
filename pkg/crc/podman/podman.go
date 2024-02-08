package podman

import (
	"fmt"
	"path/filepath"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	crcos "github.com/crc-org/crc/v2/pkg/os"
)

const (
	rootlessConn = "crc"
	rootfulConn  = "crc-root"
)

var podmanExecutablePath = filepath.Join(constants.CrcPodmanBinDir, constants.PodmanRemoteExecutableName)

func run(args ...string) (string, string, error) {
	return crcos.RunWithDefaultLocale(podmanExecutablePath, args...)
}

func addSystemConnection(name, identity, uri string) error {
	if _, stderr, err := run("system", "connection", "add", "--identity", identity, name, uri); err != nil {
		return fmt.Errorf("failed to add podman system connection %v: %s", err, stderr)
	}
	return nil
}

func AddRootlessSystemConnection(sshIdentity, uri string) error {
	return addSystemConnection(rootlessConn, sshIdentity, uri)
}

func AddRootfulSystemConnection(sshIdentity, uri string) error {
	return addSystemConnection(rootfulConn, sshIdentity, uri)
}

func removeSystemConnection(name string) error {
	if _, stderr, err := run("system", "connection", "remove", name); err != nil {
		return fmt.Errorf("failed to remove podman system connection %v: %s", err, stderr)
	}
	return nil
}

func RemoveRootlessSystemConnection() error {
	return removeSystemConnection(rootlessConn)
}

func RemoveRootfulSystemConnection() error {
	return removeSystemConnection(rootfulConn)
}

func makeSystemConnectionDefault(name string) error {
	if _, stderr, err := run("system", "connection", "default", name); err != nil {
		return fmt.Errorf("failed to make %s podman system connection as default %v: %s", name, err, stderr)
	}
	return nil
}

func MakeRootlessSystemConnectionDefault() error {
	return makeSystemConnectionDefault(rootlessConn)
}

func MakeRootfulSystemConnectionDefault() error {
	return makeSystemConnectionDefault(rootfulConn)
}
