package network

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateProxyURL(t *testing.T) {
	assert.NoError(t, ValidateProxyURL("http://company.com"))

	assert.EqualError(t, ValidateProxyURL("company.com:8080"), "Proxy URL 'company.com:8080' is not valid: url should start with http://")
	assert.EqualError(t, ValidateProxyURL("https://company.com"), "Proxy URL 'https://company.com' is not valid: https is not supported")
}

func TestHidePassword(t *testing.T) {
	url, err := URIStringForDisplay("https://user:secret@proxy.org:123")
	assert.NoError(t, err)
	assert.Equal(t, "https://user:xxx@proxy.org:123", url)
}

func TestMarshal(t *testing.T) {
	bin, err := json.Marshal(&ProxyConfig{
		HTTPProxy:   "http://proxy.org",
		HTTPSProxy:  "http://proxy.org",
		noProxy:     defaultNoProxies,
		ProxyCAFile: "/home/john/ca.crt",
		ProxyCACert: "very long string",
	})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"HTTPProxy":"http://proxy.org", "HTTPSProxy":"http://proxy.org", "ProxyCACert":"very long string", "ProxyCAFile":"/home/john/ca.crt"}`, string(bin))
}
