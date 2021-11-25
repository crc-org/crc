package preset

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/logging"
)

type Preset string

const (
	Podman    Preset = "podman"
	OpenShift Preset = "openshift"
)

func (preset Preset) String() string {
	switch preset {
	case Podman:
		return string(Podman)
	case OpenShift:
		return string(OpenShift)
	}
	return "invalid"
}

func ParsePresetE(input string) (Preset, error) {
	switch input {
	case Podman.String():
		return Podman, nil
	case OpenShift.String():
		return OpenShift, nil
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
