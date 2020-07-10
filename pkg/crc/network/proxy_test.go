package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateProxyURL(t *testing.T) {
	assert.NoError(t, ValidateProxyURL("http://company.com"))

	assert.EqualError(t, ValidateProxyURL("company.com:8080"), "Proxy URL 'company.com:8080' is not valid: url should start with http://")
	assert.EqualError(t, ValidateProxyURL("https://company.com"), "Proxy URL 'https://company.com' is not valid: https is not supported")
}
