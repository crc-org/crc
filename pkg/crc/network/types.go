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

func ParseMode(input string) Mode {
	switch input {
	case string(VSockMode):
		return VSockMode
	case string(DefaultMode):
		return DefaultMode
	default:
		logging.Errorf("unexpected network mode %s, using default", input)
		return DefaultMode
	}
}

func ValidateMode(val interface{}) (bool, string) {
	if cast.ToString(val) == string(DefaultMode) || cast.ToString(val) == string(VSockMode) {
		return true, ""
	}
	return false, fmt.Sprintf("network mode should be either %s or %s", DefaultMode, VSockMode)
}

func SuccessfullyAppliedMode(_ string, _ interface{}) string {
	return "Network mode changed. Please run `crc cleanup` and `crc setup`."
}
