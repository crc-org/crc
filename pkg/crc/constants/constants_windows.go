package constants

const (
	DefaultVMDriver = "hyperv"
	OcBinaryName    = "oc.exe"
	DefaultOcURL    = "https://mirror.openshift.com/pub/openshift-v4/clients/oc/latest/windows/oc.zip"
)

var (
	SupportedVMDrivers = [...]string{
		"virtualbox",
	}
)
