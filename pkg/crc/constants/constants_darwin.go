package constants

import (
	"path/filepath"
)

const (
	OcExecutableName           = "oc"
	PodmanRemoteExecutableName = "podman"
	TrayExecutableName         = "Red Hat OpenShift Local.app"
	DaemonAgentLabel           = "com.redhat.crc.daemon"
	QemuGuestAgentPort         = 1234
)

var (
	TapSocketPath        = filepath.Join(CrcBaseDir, "tap.sock")
	DaemonHTTPSocketPath = filepath.Join(CrcBaseDir, "crc-http.sock")
)
