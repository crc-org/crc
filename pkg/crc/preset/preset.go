package preset

import (
	"fmt"
	"runtime"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/version"
)

type Preset interface {
	fmt.Stringer
	BundleFilename() string
	BundleVersion() string
	PullSecretRequired() bool
	MinCPUs() int
	MinMemoryMiB() int
}

type PodmanPreset struct{}
type OpenShiftPreset struct{ OkdPreset }
type OkdPreset struct{}

var (
	Podman    = PodmanPreset{}
	OpenShift = OpenShiftPreset{}
	OKD       = OkdPreset{}
)

func (preset PodmanPreset) String() string {
	return "podman"
}

func (preset PodmanPreset) BundleVersion() string {
	return version.GetPodmanVersion()
}

func (preset PodmanPreset) BundleFilename() string {
	return fmt.Sprintf("crc_podman_%s_%s_%s.crcbundle", hypervisorForGoos(runtime.GOOS), preset.BundleVersion(), runtime.GOARCH)
}

func (preset PodmanPreset) PullSecretRequired() bool {
	return false
}

func (preset PodmanPreset) MinCPUs() int {
	return 2
}

func (preset PodmanPreset) MinMemoryMiB() int {
	return 2048
}

func (preset OpenShiftPreset) String() string {
	return "openshift"
}
func (preset OpenShiftPreset) BundleVersion() string {
	return version.GetBundleVersion()
}

func (preset OpenShiftPreset) PullSecretRequired() bool {
	return true
}

func (preset OpenShiftPreset) BundleFilename() string {
	return fmt.Sprintf("crc_%s_%s_%s.crcbundle", hypervisorForGoos(runtime.GOOS), preset.BundleVersion(), runtime.GOARCH)
}

func (preset OkdPreset) String() string {
	return "okd"
}

func (preset OkdPreset) BundleVersion() string {
	return version.GetOKDVersion()
}

func (preset OkdPreset) BundleFilename() string {
	return fmt.Sprintf("crc_%s_%s_%s.crcbundle", hypervisorForGoos(runtime.GOOS), preset.BundleVersion(), runtime.GOARCH)
}

func (preset OkdPreset) PullSecretRequired() bool {
	return true
}

func (preset OkdPreset) MinCPUs() int {
	return 4
}

func (preset OkdPreset) MinMemoryMiB() int {
	return 9216
}

func ParsePresetE(input string) (Preset, error) {
	switch input {
	case Podman.String():
		return Podman, nil
	case OpenShift.String():
		return OpenShift, nil
	case OKD.String():
		return OKD, nil
	default:
		return OpenShift, fmt.Errorf("Cannot parse preset '%s'", input)
	}
}
func ParsePreset(input string) Preset {
	preset, err := ParsePresetE(input)
	if err != nil {
		logging.Errorf("unexpected preset mode %s, using default", input)
		return OpenShift
	}
	return preset
}

func hypervisorForGoos(goos string) string {
	switch goos {
	case "darwin":
		return "vfkit"
	case "linux":
		return "libvirt"
	case "windows":
		return "hyperv"
	default:
		return ""
	}
}
