package segment

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/pborman/uuid"
	"github.com/segmentio/analytics-go"
)

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
	client, err := analytics.NewWithConfig("cvpHsNcmGCJqVzf6YxrSnVlwFSAZaYtp", analytics.Config{
		DefaultContext: &analytics.Context{
			App: analytics.AppInfo{
				Name:    constants.DefaultName,
				Version: version.GetCRCVersion(),
			},
		},
		Endpoint: segmentEndpoint,
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

func (c *Client) Upload(err error) error {
	if !c.config.Get(config.ConsentTelemetry).AsBool() {
		return nil
	}
	logging.Info("Uploading the error to segment")

	anonymousID, uerr := getUserIdentity(c.telemetryFilePath)
	if uerr != nil {
		return uerr
	}

	t := analytics.NewTraits().
		Set("error", err.Error())

	return c.segmentClient.Enqueue(analytics.Identify{
		AnonymousId: anonymousID,
		Traits:      t,
	})
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
