package constants

import (
	"path/filepath"
)

const (
	OcExecutableName   = "oc"
	DaemonAgentLabel   = "com.redhat.crc.daemon"
	QemuGuestAgentPort = 1234
)

var (
	TapSocketPath        = filepath.Join(CrcBaseDir, "tap.sock")
	DaemonHTTPSocketPath = filepath.Join(CrcBaseDir, "crc-http.sock")
	UnixgramSocketPath   = filepath.Join(CrcBaseDir, "crc-unixgram.sock")
)
