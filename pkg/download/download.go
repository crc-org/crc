package download

import (
	"context"
	"net/http"
	"os"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"

	"github.com/cavaliercoder/grab"
	"github.com/cavaliercoder/grab/grabui"
	"github.com/pkg/errors"
)

func Download(uri, destination string, mode os.FileMode) (string, error) {
	logging.Debugf("Downloading %s to %s", uri, destination)

	client := grab.NewClient()
	client.HTTPClient = &http.Client{Transport: network.HTTPTransport()}
	consoleClient := grabui.NewConsoleClient(client)
	req, err := grab.NewRequest(destination, uri)
	if err != nil {
		return "", errors.Wrapf(err, "unable to get response from %s", uri)
	}

	respCh := consoleClient.Do(context.Background(), 3, req)
	resp := <-respCh

	if resp.Err() != nil {
		return "", err
	}

	if err := os.Chmod(resp.Filename, mode); err != nil {
		_ = os.Remove(resp.Filename)
		return "", err
	}

	logging.Debugf("Download saved to %v", resp.Filename)
	return resp.Filename, nil
}
