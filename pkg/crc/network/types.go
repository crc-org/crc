package network

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/spf13/cast"
)

type Mode string

const (
	SystemNetworkingMode Mode = "system"
	UserNetworkingMode   Mode = "user"
)

func parseMode(input string) (Mode, error) {
	switch input {
	case string(UserNetworkingMode), "vsock":
		return UserNetworkingMode, nil
	case string(SystemNetworkingMode), "default":
		return SystemNetworkingMode, nil
	default:
		return SystemNetworkingMode, fmt.Errorf("Cannot parse mode '%s'", input)
	}
}
func ParseMode(input string) Mode {
	mode, err := parseMode(input)
	if err != nil {
		logging.Errorf("unexpected network mode %s, using default", input)
		return SystemNetworkingMode
	}
	return mode
}

func ValidateMode(val interface{}) (bool, string) {
	_, err := parseMode(cast.ToString(val))
	if err != nil {
		return false, fmt.Sprintf("network mode should be either %s or %s", SystemNetworkingMode, UserNetworkingMode)
	}
	return true, ""
}

func SuccessfullyAppliedMode(_ string, _ interface{}) string {
	return "Network mode changed. Please run `crc cleanup` and `crc setup`."
}
