//+build darwin build

package hyperkit

import "fmt"

const (
	MachineDriverCommand = "crc-driver-hyperkit"
	MachineDriverVersion = "0.12.7"
	HyperKitCommand      = "hyperkit"
	HyperKitVersion      = "v0.20180403-36-ge3b37f"
)

var (
	HyperKitDownloadURL      = fmt.Sprintf("https://github.com/code-ready/machine-driver-hyperkit/releases/download/v%s/hyperkit", MachineDriverVersion)
	MachineDriverDownloadURL = fmt.Sprintf("https://github.com/code-ready/machine-driver-hyperkit/releases/download/v%s/crc-driver-hyperkit", MachineDriverVersion)
)
