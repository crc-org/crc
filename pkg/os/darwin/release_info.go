//go:build darwin

package darwin

import (
	"fmt"
	"sync"

	"github.com/Masterminds/semver/v3"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/os"
	"github.com/crc-org/crc/v2/pkg/strings"
	"github.com/pkg/errors"
)

var (
	once              sync.Once
	macCurrentVersion string
	errGettingVersion error
)

func AtLeast(targetVersion string) (bool, error) {
	once.Do(func() {
		// syscall.Sysctl("kern.osproductversion") is not providing the consistent version info for intel and silicon macs
		// so using sw_vers -productVersion to get the version info consistently
		// https://github.com/golang/go/issues/58722
		macCurrentVersion, _, errGettingVersion = os.RunWithDefaultLocale("sw_vers", "-productVersion")
		macCurrentVersion = strings.TrimTrailingEOL(macCurrentVersion)
		logging.Debugf("sw_vers -productVersion is: %s", macCurrentVersion)
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
