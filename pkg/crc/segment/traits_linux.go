package segment

import (
	"os"
	"runtime"

	"github.com/crc-org/crc/v2/pkg/os/linux"
	"github.com/segmentio/analytics-go/v3"
)

func traits() analytics.Traits {
	base := analytics.NewTraits().
		Set("os", runtime.GOOS)

	details, err := linux.GetOsRelease()
	if err != nil {
		return base
	}
	return base.
		Set("os_release_name", details.Name).
		Set("os_release_version_id", details.VersionID).
		Set("lang", os.Getenv("LANG"))
}
