package network

import (
	"fmt"
	"os"
	"strings"

	"github.com/asaskevich/govalidator"
)

var (
	DefaultProxy     *ProxyConfig
	defaultNoProxies = "127.0.0.1, localhost"
)

// ProxyConfig keeps the proxy configuration for the current environment
type ProxyConfig struct {
	HttpProxy  string
	HttpsProxy string
	NoProxy    string
}

// NewProxyConfig creates a proxy configuration with the specified parameters. If an empty string is passed
// the corresponding environment variable is checked.
func NewProxyConfig() (*ProxyConfig, error) {
	if DefaultProxy.HttpProxy == "" {
		DefaultProxy.HttpProxy = getProxyFromEnv("http_proxy")
	}
	err := ValidateProxyURL(DefaultProxy.HttpProxy)
	if err != nil {
		return nil, err
	}

	if DefaultProxy.HttpsProxy == "" {
		DefaultProxy.HttpsProxy = getProxyFromEnv("https_proxy")
	}
	err = ValidateProxyURL(DefaultProxy.HttpsProxy)
	if err != nil {
		return nil, err
	}

	np := defaultNoProxies

	if DefaultProxy.NoProxy == "" {
		DefaultProxy.NoProxy = getProxyFromEnv("no_proxy")
	}

	if DefaultProxy.NoProxy != "" {
		np = fmt.Sprintf("%s,%s", np, DefaultProxy.NoProxy)
	}

	config := ProxyConfig{
		HttpProxy:  DefaultProxy.HttpProxy,
		HttpsProxy: DefaultProxy.HttpsProxy,
		NoProxy:    np,
	}

	return &config, nil
}

func getProxyFromEnv(proxyScheme string) string {
	p := os.Getenv(proxyScheme)
	if p == "" {
		p = os.Getenv(strings.ToUpper(proxyScheme))
	}
	return p
}

// HttpProxy with hidden credentials
func (p *ProxyConfig) HttpProxyForDisplay() string {
	httpProxy, _ := UriStringForDisplay(p.HttpProxy)
	return httpProxy
}

// HttpsProxy with hidden credentials
func (p *ProxyConfig) HttpsProxyForDisplay() string {
	httpsProxy, _ := UriStringForDisplay(p.HttpsProxy)
	return httpsProxy
}

// AddNoProxy appends the specified host to the list of no proxied hosts.
func (p *ProxyConfig) AddNoProxy(host ...string) {
	host = append(host, p.NoProxy)
	p.NoProxy = strings.Join(host, ",")
}

// Sets the current config as environment variables in the current process.
func (p *ProxyConfig) ApplyToEnvironment() {
	if !p.IsEnabled() {
		return
	}

	if p.HttpProxy != "" {
		os.Setenv("HTTP_PROXY", p.HttpProxy)
		os.Setenv("http_proxy", p.HttpProxy)
	}
	if p.HttpsProxy != "" {
		os.Setenv("HTTPS_PROXY", p.HttpsProxy)
		os.Setenv("https_proxy", p.HttpsProxy)
	}
	if len(p.NoProxy) != 0 {
		os.Setenv("NO_PROXY", p.NoProxy)
		os.Setenv("no_proxy", p.NoProxy)
	}
}

// Enabled returns true if at least one proxy (HTTP or HTTPS) is configured. Returns false otherwise.
func (p *ProxyConfig) IsEnabled() bool {
	return p.HttpProxy != "" || p.HttpsProxy != ""
}

// ValidateProxyURL validates that the specified proxyURL is valid
func ValidateProxyURL(proxyUrl string) error {
	if proxyUrl == "" {
		return nil
	}

	if !govalidator.IsURL(proxyUrl) {
		return fmt.Errorf("Proxy URL '%s' is not valid.", proxyUrl)
	}
	return nil
}
