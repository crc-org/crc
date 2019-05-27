package constants

const (
	DefaultVMDriver = "libvirt"
	OcBinaryName    = "oc"
	DefaultOcURL    = "https://mirror.openshift.com/pub/openshift-v4/clients/oc/latest/linux/oc.tar.gz"
)

var (
	SupportedVMDrivers = [...]string{
		"libvirt",
	}
)
