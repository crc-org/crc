package network

import (
	"fmt"
	"os"
	"strings"

	"github.com/asaskevich/govalidator"
)

var (
	DefaultProxy     ProxyConfig
	defaultNoProxies = []string{"127.0.0.1", "localhost"}
)

// ProxyConfig keeps the proxy configuration for the current environment
type ProxyConfig struct {
	HttpProxy  string
	HttpsProxy string
	noProxy    []string
}

func NewProxyDefaults(httpProxy, httpsProxy, noProxy string) (*ProxyConfig, error) {
	DefaultProxy = ProxyConfig{
		HttpProxy:  httpProxy,
		HttpsProxy: httpsProxy,
	}

	if DefaultProxy.HttpProxy == "" {
		DefaultProxy.HttpProxy = getProxyFromEnv("http_proxy")
	}
	if DefaultProxy.HttpsProxy == "" {
		DefaultProxy.HttpsProxy = getProxyFromEnv("https_proxy")
	}
	if noProxy == "" {
		noProxy = getProxyFromEnv("no_proxy")
	}
	DefaultProxy.setNoProxyString(noProxy)

	return NewProxyConfig()
}

// NewProxyConfig creates a proxy configuration with the specified parameters. If an empty string is passed
// the corresponding environment variable is checked.
func NewProxyConfig() (*ProxyConfig, error) {
	config := ProxyConfig{
		HttpProxy:  DefaultProxy.HttpProxy,
		HttpsProxy: DefaultProxy.HttpsProxy,
	}

	config.noProxy = defaultNoProxies
	if len(DefaultProxy.noProxy) != 0 {
		config.AddNoProxy(DefaultProxy.noProxy...)
	}

	err := ValidateProxyURL(config.HttpProxy)
	if err != nil {
		return nil, err
	}

	err = ValidateProxyURL(config.HttpsProxy)
	if err != nil {
		return nil, err
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
	p.noProxy = append(p.noProxy, host...)
}

func (p *ProxyConfig) setNoProxyString(noProxies string) {
	if noProxies == "" {
		return
	}
	p.noProxy = strings.Split(noProxies, ",")
}

func (p *ProxyConfig) GetNoProxyString() string {
	return strings.Join(p.noProxy, ",")
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
	if len(p.noProxy) != 0 {
		os.Setenv("NO_PROXY", p.GetNoProxyString())
		os.Setenv("no_proxy", p.GetNoProxyString())
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

	if strings.HasPrefix(proxyUrl, "https://") {
		return fmt.Errorf("Proxy URL '%s' is not valid: https is not supported", proxyUrl)
	}
	if !strings.HasPrefix(proxyUrl, "http://") {
		return fmt.Errorf("Proxy URL '%s' is not valid: url should start with http://", proxyUrl)
	}
	if !govalidator.IsURL(proxyUrl) {
		return fmt.Errorf("Proxy URL '%s' is not valid.", proxyUrl)
	}
	return nil
}
