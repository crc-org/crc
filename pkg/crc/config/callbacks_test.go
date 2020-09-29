package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccessfullyApplied(t *testing.T) {
	assert.Equal(t, "Successfully configured http-proxy to http://proxy", SuccessfullyApplied("http-proxy", "http://proxy"))
	assert.Equal(t, "Successfully configured enable-experimental-features to true", SuccessfullyApplied("enable-experimental-features", true))
}
