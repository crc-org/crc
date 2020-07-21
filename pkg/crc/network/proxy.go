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
	HTTPProxy  string
	HTTPSProxy string
	noProxy    []string
}

func NewProxyDefaults(httpProxy, httpsProxy, noProxy string) (*ProxyConfig, error) {
	DefaultProxy = ProxyConfig{
		HTTPProxy:  httpProxy,
		HTTPSProxy: httpsProxy,
	}

	if DefaultProxy.HTTPProxy == "" {
		DefaultProxy.HTTPProxy = getProxyFromEnv("http_proxy")
	}
	if DefaultProxy.HTTPSProxy == "" {
		DefaultProxy.HTTPSProxy = getProxyFromEnv("https_proxy")
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
		HTTPProxy:  DefaultProxy.HTTPProxy,
		HTTPSProxy: DefaultProxy.HTTPSProxy,
	}

	config.noProxy = defaultNoProxies
	if len(DefaultProxy.noProxy) != 0 {
		config.AddNoProxy(DefaultProxy.noProxy...)
	}

	err := ValidateProxyURL(config.HTTPProxy)
	if err != nil {
		return nil, err
	}

	err = ValidateProxyURL(config.HTTPSProxy)
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

// HTTPProxy with hidden credentials
func (p *ProxyConfig) HTTPProxyForDisplay() string {
	httpProxy, _ := URIStringForDisplay(p.HTTPProxy)
	return httpProxy
}

// HTTPSProxy with hidden credentials
func (p *ProxyConfig) HTTPSProxyForDisplay() string {
	httpsProxy, _ := URIStringForDisplay(p.HTTPSProxy)
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

	if p.HTTPProxy != "" {
		os.Setenv("HTTP_PROXY", p.HTTPProxy)
		os.Setenv("http_proxy", p.HTTPProxy)
	}
	if p.HTTPSProxy != "" {
		os.Setenv("HTTPS_PROXY", p.HTTPSProxy)
		os.Setenv("https_proxy", p.HTTPSProxy)
	}
	if len(p.noProxy) != 0 {
		os.Setenv("NO_PROXY", p.GetNoProxyString())
		os.Setenv("no_proxy", p.GetNoProxyString())
	}
}

// Enabled returns true if at least one proxy (HTTP or HTTPS) is configured. Returns false otherwise.
func (p *ProxyConfig) IsEnabled() bool {
	return p.HTTPProxy != "" || p.HTTPSProxy != ""
}

// ValidateProxyURL validates that the specified proxyURL is valid
func ValidateProxyURL(proxyURL string) error {
	if proxyURL == "" {
		return nil
	}

	if strings.HasPrefix(proxyURL, "https://") {
		return fmt.Errorf("Proxy URL '%s' is not valid: https is not supported", proxyURL)
	}
	if !strings.HasPrefix(proxyURL, "http://") {
		return fmt.Errorf("Proxy URL '%s' is not valid: url should start with http://", proxyURL)
	}
	if !govalidator.IsURL(proxyURL) {
		return fmt.Errorf("Proxy URL '%s' is not valid", proxyURL)
	}
	return nil
}
