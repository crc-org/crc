package constants

import (
	"path/filepath"
)

const (
	OcExecutableName           = "oc"
	PodmanRemoteExecutableName = "podman"
	DaemonAgentLabel           = "com.redhat.crc.daemon"
	QemuGuestAgentPort         = 1234
)

var (
	TapSocketPath        = filepath.Join(CrcBaseDir, "tap.sock")
	DaemonHTTPSocketPath = filepath.Join(CrcBaseDir, "crc-http.sock")
)
