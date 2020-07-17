package download

import (
	"fmt"
	"os"

	"github.com/cavaliercoder/grab"
	"github.com/code-ready/crc/pkg/crc/logging"
)

func Download(uri, destination string, mode os.FileMode) (string, error) {
	// create client
	logging.Debugf("Downloading %s to %s", uri, destination)
	client := grab.NewClient()
	req, err := grab.NewRequest(destination, uri)
	if err != nil {
		return "", fmt.Errorf("Unable to get response from %s: %v", uri, err)
	}
	defer func() {
		if err != nil {
			os.Remove(destination)
		}
	}()
	resp := client.Do(req)
	// check for errors
	if err := resp.Err(); err != nil {
		return "", fmt.Errorf("Download failed: %v", err)
	}

	err = os.Chmod(resp.Filename, mode)
	if err != nil {
		os.Remove(resp.Filename)
		return "", err
	}

	logging.Debugf("Download saved to %v", resp.Filename)
	return resp.Filename, nil
}
