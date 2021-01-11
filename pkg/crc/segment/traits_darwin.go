package segment

import (
	"runtime"
	"strings"

	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/segmentio/analytics-go"
)

func traits() analytics.Traits {
	base := analytics.NewTraits().
		Set("os", runtime.GOOS)

	version, _, err := crcos.RunWithDefaultLocale("sw_vers", "-productVersion")
	if err != nil {
		return base
	}
	return base.Set("os_version", strings.TrimSpace(version))
}
