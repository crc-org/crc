package host

import (
	"testing"

	"github.com/code-ready/crc/pkg/drivers/fakedriver"
	_ "github.com/code-ready/crc/pkg/drivers/none"
	"github.com/code-ready/crc/pkg/libmachine/state"
)

func TestValidateHostnameValid(t *testing.T) {
	hosts := []string{
		"zomg",
		"test-ing",
		"some.h0st",
	}

	for _, v := range hosts {
		isValid := ValidateHostName(v)
		if !isValid {
			t.Fatalf("Thought a valid hostname was invalid: %s", v)
		}
	}
}

func TestValidateHostnameInvalid(t *testing.T) {
	hosts := []string{
		"zom_g",
		"test$ing",
		"some😄host",
	}

	for _, v := range hosts {
		isValid := ValidateHostName(v)
		if isValid {
			t.Fatalf("Thought an invalid hostname was valid: %s", v)
		}
	}
}

func TestStart(t *testing.T) {
	host := &Host{
		Driver: &fakedriver.Driver{
			MockState: state.Stopped,
		},
	}

	if err := host.Start(); err != nil {
		t.Fatalf("Expected no error but got one: %s", err)
	}
}
