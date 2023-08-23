package httpproxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateProxyURL(t *testing.T) {
	assert.NoError(t, ValidateProxyURL("http://company.com", false))
	assert.NoError(t, ValidateProxyURL("http://company.com", true))
	assert.NoError(t, ValidateProxyURL("https://company.com", true))

	assert.EqualError(t, ValidateProxyURL("company.com:8080", false), "HTTP proxy URL 'company.com:8080' is not valid: url should start with http://")
	assert.EqualError(t, ValidateProxyURL("company.com:8080", true), "HTTPS proxy URL 'company.com:8080' is not valid: url should start with http:// or https://")
	assert.EqualError(t, ValidateProxyURL("https://company.com", false), "HTTP proxy URL 'https://company.com' is not valid: url should start with http://")
}
func TestTrimTrailingEOL(t *testing.T) {
	assert.Equal(t, "foo\nbar", trimTrailingEOL("foo\nbar\n"))
	assert.Equal(t, "foo", trimTrailingEOL("foo\n"))
	assert.Equal(t, "foo", trimTrailingEOL("foo\r\n"))
	assert.Equal(t, "foo\r\nbar", trimTrailingEOL("foo\r\nbar\r\n"))
	assert.Equal(t, "foo\r\nbar", trimTrailingEOL("foo\r\nbar\r\n\r\n"))
	assert.Equal(t, "foo\nbar", trimTrailingEOL("foo\nbar\n\n"))
	assert.Equal(t, "foo\nbar", trimTrailingEOL("foo\nbar\n\n\n"))
	assert.Equal(t, "", trimTrailingEOL("\r\n"))
	assert.Equal(t, "", trimTrailingEOL("\n"))
	assert.Equal(t, "", trimTrailingEOL(""))
}
