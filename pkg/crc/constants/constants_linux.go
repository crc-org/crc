package constants

import (
	"path/filepath"
)

const (
	OcExecutableName = "oc"
	TapSocketPath    = ""
)

var DaemonHTTPSocketPath = filepath.Join(CrcBaseDir, "crc-http.sock")
