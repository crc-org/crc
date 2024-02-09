//go:build darwin

package darwin

import (
	"fmt"
	"sync"
	"syscall"

	"github.com/Masterminds/semver/v3"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/pkg/errors"
)

var (
	once              sync.Once
	macCurrentVersion string
	errGettingVersion error
)

func AtLeast(targetVersion string) (bool, error) {
	once.Do(func() {
		macCurrentVersion, errGettingVersion = syscall.Sysctl("kern.osproductversion")
		logging.Debugf("kern.osproductversion is: %s", macCurrentVersion)
	})
	if errGettingVersion != nil {
		return false, errGettingVersion
	}

	cVersion, err := semver.NewVersion(macCurrentVersion)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("cannot parse %s", macCurrentVersion))
	}
	targetVersionStr, err := semver.NewVersion(targetVersion)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("cannot parse %s", targetVersion))
	}
	return cVersion.Equal(targetVersionStr) || cVersion.GreaterThan(targetVersionStr), nil
}
