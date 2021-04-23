package segment

import (
	"context"
	"errors"
	"fmt"
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
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/pborman/uuid"
	"github.com/segmentio/analytics-go"
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

func (c *Client) UploadAction(action, source string) error {
	return c.upload(action, baseProperties(source))
}

func (c *Client) UploadCmd(ctx context.Context, action string, duration time.Duration, err error) error {
	return c.upload(action, properties(ctx, err, duration))
}

func (c *Client) upload(action string, a analytics.Properties) error {
	if c.config.Get(config.ConsentTelemetry).AsString() != "yes" {
		return nil
	}

	userID, uerr := getUserIdentity(c.telemetryFilePath)
	if uerr != nil {
		return uerr
	}

	if err := c.segmentClient.Enqueue(analytics.Identify{
		UserId: userID,
		Traits: addConfigTraits(c.config, traits()),
	}); err != nil {
		return err
	}

	return c.segmentClient.Enqueue(analytics.Track{
		UserId:     userID,
		Event:      action,
		Properties: a,
	})
}

func baseProperties(source string) analytics.Properties {
	return analytics.NewProperties().
		Set("version", version.GetCRCVersion()).
		Set("source", source)
}

func properties(ctx context.Context, err error, duration time.Duration) analytics.Properties {
	properties := baseProperties("cli")
	for k, v := range telemetry.GetContextProperties(ctx) {
		properties = properties.Set(k, v)
	}

	properties = properties.
		Set("success", err == nil).
		Set("duration", duration.Milliseconds()).
		Set("tty", crcos.RunningInTerminal()).
		Set("remote", crcos.RunningUsingSSH())
	if err != nil {
		properties = properties.Set("error", telemetry.SetError(err)).
			Set("error-type", errorType(err))
	}
	return properties
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

func errorType(err error) string {
	wrappedErr := errors.Unwrap(err)
	if wrappedErr != nil {
		return fmt.Sprintf("%T", wrappedErr)
	}
	return fmt.Sprintf("%T", err)
}
