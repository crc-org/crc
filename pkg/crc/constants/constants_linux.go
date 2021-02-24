package constants

import (
	"path/filepath"
)

const (
	OcExecutableName          = "oc"
	PodmanExecutableName      = "podman"
	AdminHelperExecutableName = "admin-helper-linux"
	TapSocketPath             = ""
)

var DaemonHTTPSocketPath = filepath.Join(CrcBaseDir, "crc-http.sock")
