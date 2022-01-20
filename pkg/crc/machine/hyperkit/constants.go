//go:build darwin || build
// +build darwin build

package hyperkit

import "fmt"

const (
	MachineDriverCommand = "crc-driver-hyperkit"
	MachineDriverVersion = "0.15.1"
	HyperKitCommand      = "hyperkit"
	HyperKitVersion      = "v0.20210107-12-gadc4ea"
	QcowToolCommand      = "qcow-tool"
	QcowToolVersion      = "1.0.0"
)

var (
	baseURL = fmt.Sprintf("https://github.com/code-ready/machine-driver-hyperkit/releases/download/v%s", MachineDriverVersion)

	HyperKitDownloadURL      = fmt.Sprintf("%s/%s", baseURL, HyperKitCommand)
	MachineDriverDownloadURL = fmt.Sprintf("%s/%s", baseURL, MachineDriverCommand)
	QcowToolDownloadURL      = fmt.Sprintf("%s/%s", baseURL, QcowToolCommand)
)
