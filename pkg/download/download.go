package download

import (
	"github.com/code-ready/crc/pkg/crc/logging"

	"github.com/cavaliercoder/grab"
	"github.com/code-ready/crc/pkg/crc/errors"
)

func Download(uri, destination string) (string, error) {
	// create client
	client := grab.NewClient()
	req, err := grab.NewRequest(destination, uri)
	if err != nil {
		return "", errors.NewF("Not able to get response from %s: %v", uri, err)
	}
	resp := client.Do(req)
	// check for errors
	if err := resp.Err(); err != nil {
		return "", errors.NewF("Download failed: %v\n", err)
	}

	logging.DebugF("Download saved to ./%v \n", resp.Filename)
	return resp.Filename, nil
}
