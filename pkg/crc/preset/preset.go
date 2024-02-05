package preset

import (
	"fmt"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
)

type Preset string

const (
	Podman     Preset = "podman"
	OpenShift  Preset = "openshift"
	OKD        Preset = "okd"
	Microshift Preset = "microshift"
)

var presetMap = map[Preset]string{
	Podman:     string(Podman),
	OpenShift:  string(OpenShift),
	OKD:        string(OKD),
	Microshift: string(Microshift),
}

const (
	PodmanDeprecatedWarning = "The Podman preset is deprecated and will be removed in a future release. Consider" +
		" rather using a Podman Machine managed by Podman Desktop: https://podman-desktop.io"
)

func AllPresets() []Preset {
	var keys []Preset
	for k := range presetMap {
		keys = append(keys, k)
	}
	return keys
}

func (preset Preset) String() string {
	presetStr, presetExists := presetMap[preset]
	if !presetExists {
		return "invalid"
	}
	return presetStr
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
	for pSet, pString := range presetMap {
		if pString == input {
			return pSet, nil
		}
	}
	return OpenShift, fmt.Errorf("Cannot parse preset '%s'", input)

}
func ParsePreset(input string) Preset {
	preset, err := ParsePresetE(input)
	if err != nil {
		logging.Errorf("unexpected preset mode %s, using default", input)
		return OpenShift
	}
	return preset
}
