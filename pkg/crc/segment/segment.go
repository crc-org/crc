package segment

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/telemetry"
	"github.com/code-ready/crc/pkg/crc/version"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/pborman/uuid"
	"github.com/segmentio/analytics-go"
	"github.com/spf13/cast"
)

var WriteKey = "R7jGNYYO5gH0Nl5gDlMEuZ3gPlDJKQak" // test

type Client struct {
	segmentClient analytics.Client
	config        *crcConfig.Config
	userID        string
	identifyHash  uint64
}

func NewClient(config *crcConfig.Config, transport http.RoundTripper) (*Client, error) {
	return newCustomClient(config, transport,
		filepath.Join(constants.GetHomeDir(), ".redhat", "anonymousId"),
		analytics.DefaultEndpoint)
}

func newCustomClient(config *crcConfig.Config, transport http.RoundTripper, telemetryFilePath, segmentEndpoint string) (*Client, error) {
	userID, err := getUserIdentity(telemetryFilePath)
	if err != nil {
		return nil, err
	}
	client, err := analytics.NewWithConfig(WriteKey, analytics.Config{
		Endpoint: segmentEndpoint,
		Logger:   &loggingAdapter{},
		DefaultContext: &analytics.Context{
			IP: net.IPv4(0, 0, 0, 0),
		},
		Transport: transport,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		segmentClient: client,
		config:        config,
		userID:        userID,
	}, nil
}

func (c *Client) Close() error {
	return c.segmentClient.Close()
}

func (c *Client) UploadAction(action, source, status string) error {
	return c.upload(action, baseProperties(source).
		Set("status", status))
}

func (c *Client) UploadCmd(ctx context.Context, action string, duration time.Duration, err error) error {
	return c.upload(action, properties(ctx, err, duration))
}

func identifyHash(identify *analytics.Identify) (uint64, error) {
	h := fnv.New64a()
	_, err := io.WriteString(h, identify.UserId)
	if err != nil {
		logging.Warnf("Failed to calculate checksum for userID '%s'", identify.UserId)
		return 0, err
	}

	// Make sure we hash the map fields always in the same order
	traits := []string{}
	for trait := range identify.Traits {
		traits = append(traits, trait)
	}
	sort.Strings(traits)
	for _, trait := range traits {
		_, err := io.WriteString(h, trait)
		if err != nil {
			logging.Warnf("Failed to calculate checksum for '%s'", trait)
			return 0, err
		}
		str, err := cast.ToStringE(identify.Traits[trait])
		if err != nil {
			logging.Warnf("Could not convert 'Traits[%s]' to string, extraneous 'identify' may be sent to segment", trait)
			return 0, err
		}
		_, err = io.WriteString(h, str)
		if err != nil {
			logging.Warnf("Failed to calculate checksum for '%s'", str)
			return 0, err
		}
	}

	return h.Sum64(), nil
}

func (c *Client) identify() *analytics.Identify {
	return &analytics.Identify{
		UserId: c.userID,
		Traits: addConfigTraits(c.config, traits()),
	}
}

func (c *Client) upload(action string, a analytics.Properties) error {
	if c.config.Get(crcConfig.ConsentTelemetry).AsString() != "yes" {
		return nil
	}

	identify := c.identify()
	hash, err := identifyHash(identify)
	if err != nil || hash != c.identifyHash {
		logging.Debug("Sending 'identify' to segment")
		if err := c.segmentClient.Enqueue(identify); err != nil {
			return err
		}
		c.identifyHash = hash
	}

	return c.segmentClient.Enqueue(analytics.Track{
		UserId:     c.userID,
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
		Set(crcConfig.NetworkMode, crcConfig.GetNetworkMode(c)).
		Set(crcConfig.EnableClusterMonitoring, c.Get(crcConfig.EnableClusterMonitoring).AsBool()).
		Set(crcConfig.ExperimentalFeatures, c.Get(crcConfig.ExperimentalFeatures).AsBool())
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
