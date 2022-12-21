package daemonclient

import (
	"fmt"
	"net/http"

	networkclient "github.com/containers/gvisor-tap-vsock/pkg/client"
	"github.com/crc-org/crc/pkg/crc/api/client"
	crcversion "github.com/crc-org/crc/pkg/crc/version"
	pkgerrors "github.com/pkg/errors"
)

const genericDaemonNotRunningMessage = "Is 'crc daemon' running? Cannot reach daemon API"

type Client struct {
	NetworkClient   *networkclient.Client
	APIClient       client.Client
	WebSocketClient *client.WebSocketClient
}

func New() *Client {
	return &Client{
		NetworkClient: networkclient.New(&http.Client{
			Transport: transport(),
		}, "http://unix/network"),
		APIClient: client.New(&http.Client{
			Transport: transport(),
		}, "http://unix/api"),
		WebSocketClient: client.NewWebSocketClient(&http.Client{
			Transport: transport(),
		}),
	}
}

func GetVersionFromDaemonAPI() (*client.VersionResult, error) {
	apiClient := client.New(&http.Client{Transport: transport()}, "http://unix/api")
	version, err := apiClient.Version()
	if err != nil {
		return nil, pkgerrors.Wrap(err, genericDaemonNotRunningMessage)
	}
	return &version, nil
}

func CheckIfOlderVersion(version *client.VersionResult) error {
	if version.CrcVersion != crcversion.GetCRCVersion() {
		return fmt.Errorf("The executable version (%s) doesn't match the daemon version (%s)", crcversion.GetCRCVersion(), version.CrcVersion)
	}
	return nil
}
