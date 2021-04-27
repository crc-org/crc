package constants

import (
	"path/filepath"
)

const (
	OcExecutableName          = "oc"
	AdminHelperExecutableName = "admin-helper-linux"
	TapSocketPath             = ""
)

var DaemonHTTPSocketPath = filepath.Join(CrcBaseDir, "crc-http.sock")
