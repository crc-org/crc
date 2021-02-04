package segment

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/telemetry"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/pborman/uuid"
	"github.com/segmentio/analytics-go"
	terminal "golang.org/x/term"
)

var WriteKey = "R7jGNYYO5gH0Nl5gDlMEuZ3gPlDJKQak" // test

type Client struct {
	segmentClient     analytics.Client
	config            *crcConfig.Config
	telemetryFilePath string
}

func NewClient(config *crcConfig.Config) (*Client, error) {
	return newCustomClient(config,
		filepath.Join(constants.GetHomeDir(), ".redhat", "anonymousId"),
		analytics.DefaultEndpoint)
}

func newCustomClient(config *crcConfig.Config, telemetryFilePath, segmentEndpoint string) (*Client, error) {
	client, err := analytics.NewWithConfig(WriteKey, analytics.Config{
		Endpoint: segmentEndpoint,
		Logger:   &loggingAdapter{},
		DefaultContext: &analytics.Context{
			IP: net.IPv4(0, 0, 0, 0),
		},
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		segmentClient:     client,
		config:            config,
		telemetryFilePath: telemetryFilePath,
	}, nil
}

func (c *Client) Close() error {
	return c.segmentClient.Close()
}

func (c *Client) Upload(ctx context.Context, action string, duration time.Duration, err error) error {
	if c.config.Get(config.ConsentTelemetry).AsString() != "yes" {
		return nil
	}

	anonymousID, uerr := getUserIdentity(c.telemetryFilePath)
	if uerr != nil {
		return uerr
	}

	if err := c.segmentClient.Enqueue(analytics.Identify{
		AnonymousId: anonymousID,
		Traits:      addConfigTraits(c.config, traits()),
	}); err != nil {
		return err
	}

	properties := analytics.NewProperties()
	for k, v := range telemetry.GetContextProperties(ctx) {
		properties = properties.Set(k, v)
	}

	properties = properties.Set("version", version.GetCRCVersion()).
		Set("success", err == nil).
		Set("duration", duration.Milliseconds()).
		Set("tty", terminal.IsTerminal(int(os.Stdin.Fd())))
	if err != nil {
		properties = properties.Set("error", err.Error())
	}

	return c.segmentClient.Enqueue(analytics.Track{
		AnonymousId: anonymousID,
		Event:       action,
		Properties:  properties,
	})
}

func addConfigTraits(c *crcConfig.Config, in analytics.Traits) analytics.Traits {
	return in.
		Set("proxy", isProxyUsed()).
		Set(config.NetworkMode, c.Get(config.NetworkMode).AsString()).
		Set(config.EnableClusterMonitoring, c.Get(config.EnableClusterMonitoring).AsBool()).
		Set(config.ExperimentalFeatures, c.Get(config.ExperimentalFeatures).AsBool())
}

func isProxyUsed() bool {
	proxyConfig, err := network.NewProxyConfig()
	if err != nil {
		return false
	}
	return proxyConfig.IsEnabled()
}

func getUserIdentity(telemetryFilePath string) (string, error) {
	var id []byte
	if err := os.MkdirAll(filepath.Dir(telemetryFilePath), 0750); err != nil {
		return "", err
	}
	if _, err := os.Stat(telemetryFilePath); !os.IsNotExist(err) {
		id, err = ioutil.ReadFile(telemetryFilePath)
		if err != nil {
			return "", err
		}
	}
	if uuid.Parse(strings.TrimSpace(string(id))) == nil {
		id = []byte(uuid.NewRandom().String())
		if err := ioutil.WriteFile(telemetryFilePath, id, 0600); err != nil {
			return "", err
		}
	}
	return strings.TrimSpace(string(id)), nil
}

type loggingAdapter struct{}

func (l *loggingAdapter) Logf(format string, args ...interface{}) {
	logging.Infof(format, args...)
}

func (l *loggingAdapter) Errorf(format string, args ...interface{}) {
	logging.Errorf(format, args...)
}
