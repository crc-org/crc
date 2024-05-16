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
