//+build darwin build

package hyperkit

import "fmt"

const (
	MachineDriverCommand = "crc-driver-hyperkit"
	MachineDriverVersion = "0.12.10-beta1"
	HyperKitCommand      = "hyperkit"
	HyperKitVersion      = "v0.20200224-44-gb54460"
)

var (
	HyperKitDownloadURL      = fmt.Sprintf("https://github.com/code-ready/machine-driver-hyperkit/releases/download/v%s/hyperkit", MachineDriverVersion)
	MachineDriverDownloadURL = fmt.Sprintf("https://github.com/code-ready/machine-driver-hyperkit/releases/download/v%s/crc-driver-hyperkit", MachineDriverVersion)
)
