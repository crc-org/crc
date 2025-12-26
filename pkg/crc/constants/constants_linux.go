package constants

import (
	"path/filepath"
)

const (
	OcExecutableName           = "oc"
	PodmanRemoteExecutableName = "podman-remote"
	TapSocketPath              = ""
	KrunkitCommand             = ""
)

var DaemonHTTPSocketPath = filepath.Join(CrcBaseDir, "crc-http.sock")

func Provider() string {
	return ""
}
