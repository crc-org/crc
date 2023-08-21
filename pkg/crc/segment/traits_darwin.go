package segment

import (
	"runtime"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/version"
	crcos "github.com/crc-org/crc/v2/pkg/os"
	"github.com/segmentio/analytics-go/v3"
)

func traits() analytics.Traits {
	base := analytics.NewTraits().
		Set("os", runtime.GOOS).
		Set("used_installer", version.IsInstaller())

	version, _, err := crcos.RunWithDefaultLocale("sw_vers", "-productVersion")
	if err != nil {
		return base
	}
	return base.Set("os_version", strings.TrimSpace(version))
}
