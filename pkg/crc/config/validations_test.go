package config

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePresetWithVariousPresetValues(t *testing.T) {
	tests := []struct {
		presetValue       string
		validationPassed  bool
		validationMessage string
	}{
		{"openshift", true, ""},
		{"microshift", true, ""},
		{"unknown", false, "Unknown preset"},
	}
	for _, tt := range tests {
		t.Run(tt.presetValue, func(t *testing.T) {
			// When
			validationPass, validationMessage := validatePreset(tt.presetValue)

			// Then
			assert.Equal(t, tt.validationPassed, validationPass)
			assert.Contains(t, validationMessage, tt.validationMessage)
		})
	}
}

func TestValidationPreset_WhenOKDProvidedOnNonArmArchitecture_thenValidationSuccessful(t *testing.T) {
	// Given
	if runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" {
		t.Skip("Skipping test; running on ARM architecture")
	}
	// When
	validationPass, validationMessage := validatePreset("okd")
	// Then
	assert.Equal(t, true, validationPass)
	assert.Equal(t, "", validationMessage)
}

func TestValidationPreset_WhenOKDProvidedOnArmArchitecture_thenValidationFailure(t *testing.T) {
	// Given
	if runtime.GOARCH != "arm" && runtime.GOARCH != "arm64" {
		t.Skip("Skipping test; not running on ARM architecture")
	}
	// When
	validationPass, validationMessage := validatePreset("okd")
	// Then
	assert.Equal(t, false, validationPass)
	assert.Equal(t, fmt.Sprintf("preset 'okd' is not supported on %s architecture, please use different preset value", runtime.GOARCH), validationMessage)
}
