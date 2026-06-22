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
	TapSocketPath        = filepath.Join(SocketBaseDir, "tap.sock")
	DaemonHTTPSocketPath = filepath.Join(SocketBaseDir, "crc-http.sock")
	UnixgramSocketPath   = filepath.Join(SocketBaseDir, "crc-unixgram.sock")
)
