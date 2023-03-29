package preset

import (
	"fmt"

	"github.com/crc-org/crc/pkg/crc/logging"
)

type Preset string

const (
	Podman     Preset = "podman"
	OpenShift  Preset = "openshift"
	OKD        Preset = "okd"
	Microshift Preset = "microshift"
)

func (preset Preset) String() string {
	switch preset {
	case Podman:
		return string(Podman)
	case OpenShift:
		return string(OpenShift)
	case OKD:
		return string(OKD)
	case Microshift:
		return string(Microshift)
	}
	return "invalid"
}

func (preset Preset) ForDisplay() string {
	switch preset {
	case Podman:
		return "Podman"
	case OpenShift:
		return "OpenShift"
	case OKD:
		return "OKD"
	case Microshift:
		return "MicroShift"
	}
	return "unknown"
}

func ParsePresetE(input string) (Preset, error) {
	switch input {
	case Podman.String():
		return Podman, nil
	case OpenShift.String():
		return OpenShift, nil
	case OKD.String():
		return OKD, nil
	case Microshift.String():
		return Microshift, nil
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
