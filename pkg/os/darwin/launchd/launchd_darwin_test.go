package launchd

import (
	"fmt"
	goos "os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testAgentConfig = AgentConfig{
	Label:          "crc.daemon",
	ExecutablePath: "/usr/local/bin/crc",
	StdOutFilePath: "/tmp/crcd.log",
	StdErrFilePath: "/tmp/crcd.log",
	Args:           []string{"daemon", "--log-level=debug"},
	Env:            map[string]string{"Version": "1.0.0"},
}

func TestGeneratePlistContentForSession(t *testing.T) {
	for _, sessionType := range []string{aquaSessionType, backgroundSessionType} {
		t.Run(sessionType, func(t *testing.T) {
			plist, err := generatePlistContentForSession(testAgentConfig, sessionType)
			require.NoError(t, err)
			assert.Contains(t, string(plist), "<key>Label</key>\n\t<string>crc.daemon</string>")
			assert.Contains(t, string(plist), "<string>/usr/local/bin/crc</string>")
			assert.Contains(t, string(plist), "<key>Version</key>\n\t\t<string>1.0.0</string>")
			assert.Contains(t, string(plist), fmt.Sprintf("<key>LimitLoadToSessionType</key>\n\t<string>%s</string>", sessionType))
		})
	}
}

func TestDomainForSessionType(t *testing.T) {
	uid := goos.Getuid()
	assert.Equal(t, fmt.Sprintf("gui/%d", uid), domainForSessionType(aquaSessionType))
	assert.Equal(t, fmt.Sprintf("user/%d", uid), domainForSessionType(backgroundSessionType))
}
