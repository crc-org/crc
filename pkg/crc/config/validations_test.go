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

func TestValidateHTTPProxy(t *testing.T) {
	tests := []struct {
		name                     string
		noProxyValue             string
		expectedValidationResult bool
		expectedValidationStr    string
	}{
		{"empty value", "", true, ""},
		{"valid http url", "http://proxy.example.com", true, ""},
		{"valid https url", "https://proxy.example.com", false, "HTTP proxy URL 'https://proxy.example.com' is not valid: url should start with http://"},
		{"valid socks5 url", "socks5://proxy.example.com", false, "HTTP proxy URL 'socks5://proxy.example.com' is not valid: url should start with http://"},
		{"type in http scheme", "htp://proxy.example.com", false, "HTTP proxy URL 'htp://proxy.example.com' is not valid: url should start with http://"},
		{"no scheme", "proxy.example.com", false, "HTTP proxy URL 'proxy.example.com' is not valid: url should start with http://"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualValidationResult, actualValidationStr := validateHTTPProxy(tt.noProxyValue)
			if actualValidationStr != tt.expectedValidationStr {
				t.Errorf("validateHTTPProxy(%s) : got %v, want %v", tt.noProxyValue, actualValidationStr, tt.expectedValidationStr)
			}
			if actualValidationResult != tt.expectedValidationResult {
				t.Errorf("validateHTTPProxy(%s) : got %v, want %v", tt.noProxyValue, actualValidationResult, tt.expectedValidationResult)
			}
		})
	}
}

func TestValidateHTTPSProxy(t *testing.T) {
	tests := []struct {
		name                     string
		noProxyValue             string
		expectedValidationResult bool
		expectedValidationStr    string
	}{
		{"empty value", "", true, ""},
		{"valid https url", "https://proxy.example.com", true, ""},
		{"valid http url", "http://proxy.example.com", true, ""},
		{"valid socks5 url", "socks5://proxy.example.com", false, "HTTPS proxy URL 'socks5://proxy.example.com' is not valid: url should start with http:// or https://"},
		{"type in https scheme", "htps://proxy.example.com", false, "HTTPS proxy URL 'htps://proxy.example.com' is not valid: url should start with http:// or https://"},
		{"no scheme", "proxy.example.com", false, "HTTPS proxy URL 'proxy.example.com' is not valid: url should start with http:// or https://"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualValidationResult, actualValidationStr := validateHTTPSProxy(tt.noProxyValue)
			if actualValidationStr != tt.expectedValidationStr {
				t.Errorf("validateHTTPSProxy(%s) : got %v, want %v", tt.noProxyValue, actualValidationStr, tt.expectedValidationStr)
			}
			if actualValidationResult != tt.expectedValidationResult {
				t.Errorf("validateHTTPSProxy(%s) : got %v, want %v", tt.noProxyValue, actualValidationResult, tt.expectedValidationResult)
			}
		})
	}
}

func TestValidateNoProxy(t *testing.T) {
	tests := []struct {
		name                     string
		noProxyValue             string
		expectedValidationResult bool
		expectedValidationStr    string
	}{
		{"empty value", "", true, ""},
		{"valid single", "example.com", true, ""},
		{"valid multiple", "localhost,127.0.0.1,example.com", true, ""},
		{"space in single entry", "example .com", false, "NoProxy string can't contain spaces"},
		{"space in between multiple entries", "localhost, , example.com", false, "NoProxy string can't contain spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualValidationResult, actualValidationStr := validateNoProxy(tt.noProxyValue)
			if actualValidationStr != tt.expectedValidationStr {
				t.Errorf("validateNoProxy(%s) : got %v, want %v", tt.noProxyValue, actualValidationStr, tt.expectedValidationStr)
			}
			if actualValidationResult != tt.expectedValidationResult {
				t.Errorf("validateNoProxy(%s) : got %v, want %v", tt.noProxyValue, actualValidationResult, tt.expectedValidationResult)
			}
		})
	}
}

func TestValidatePersistentVolumeSize(t *testing.T) {
	tests := []struct {
		name                     string
		persistentVolumeSize     string
		expectedValidationResult bool
	}{
		{"equal to 15G", "15", true},
		{"greater than 15G", "20", true},
		{"less than 15G", "7", false},
		{"invalid integer value", "an-elephant", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualValidationResult, _ := validatePersistentVolumeSize(tt.persistentVolumeSize)
			if actualValidationResult != tt.expectedValidationResult {
				t.Errorf("validatePersistentVolumeSize(%s) : got %v, want %v", tt.persistentVolumeSize, actualValidationResult, tt.expectedValidationResult)
			}
		})
	}
}
