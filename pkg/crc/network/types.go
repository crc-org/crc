package network

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/spf13/cast"
)

type NameServer struct {
	IPAddress string
}

type SearchDomain struct {
	Domain string
}

type ResolvFileValues struct {
	SearchDomains []SearchDomain
	NameServers   []NameServer
}

type Mode string

const (
	DefaultMode Mode = "default"
	VSockMode   Mode = "vsock"
)

func parseMode(input string) (Mode, error) {
	switch input {
	case string(VSockMode):
		return VSockMode, nil
	case string(DefaultMode):
		return DefaultMode, nil
	default:
		return DefaultMode, fmt.Errorf("Cannot parse mode '%s'", input)
	}
}
func ParseMode(input string) Mode {
	mode, err := parseMode(input)
	if err != nil {
		logging.Errorf("unexpected network mode %s, using default", input)
		return DefaultMode
	}
	return mode
}

func ValidateMode(val interface{}) (bool, string) {
	_, err := parseMode(cast.ToString(val))
	if err != nil {
		return false, fmt.Sprintf("network mode should be either %s or %s", DefaultMode, VSockMode)
	}
	return true, ""
}

func SuccessfullyAppliedMode(_ string, _ interface{}) string {
	return "Network mode changed. Please run `crc cleanup` and `crc setup`."
}
