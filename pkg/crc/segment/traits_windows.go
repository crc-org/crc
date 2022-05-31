package segment

import (
	"runtime"
	"strings"

	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/code-ready/crc/pkg/os/windows/powershell"
	"github.com/segmentio/analytics-go/v3"
)

func traits() analytics.Traits {
	base := analytics.NewTraits().
		Set("os", runtime.GOOS).
		Set("used_installer", version.IsInstaller())

	releaseID, _, err := powershell.Execute(`(Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion" -Name ReleaseId).ReleaseId`)
	if err != nil {
		return base
	}
	editionID, _, err := powershell.Execute(`(Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion").EditionID`)
	if err != nil {
		return base
	}
	return base.
		Set("os_edition_id", strings.TrimSpace(editionID)).
		Set("os_release_id", strings.TrimSpace(releaseID))
}
