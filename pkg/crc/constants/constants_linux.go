package constants

import (
	"path/filepath"
)

const (
	OcExecutableName           = "oc"
	PodmanRemoteExecutableName = "podman-remote"
	TapSocketPath              = ""
	UnixgramSocketPath         = ""
)

var DaemonHTTPSocketPath = filepath.Join(CrcBaseDir, "crc-http.sock")
