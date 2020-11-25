package download

import (
	"os"

	"github.com/cavaliercoder/grab"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/pkg/errors"
)

func Download(uri, destination string, mode os.FileMode) (string, error) {
	logging.Debugf("Downloading %s to %s", uri, destination)

	client := grab.NewClient()
	req, err := grab.NewRequest(destination, uri)
	if err != nil {
		return "", errors.Wrapf(err, "unable to get response from %s", uri)
	}

	resp := client.Do(req)
	if err := resp.Err(); err != nil {
		return "", errors.Wrapf(err, "download of %s failed", uri)
	}

	if err := os.Chmod(resp.Filename, mode); err != nil {
		_ = os.Remove(resp.Filename)
		return "", err
	}

	logging.Debugf("Download saved to %v", resp.Filename)
	return resp.Filename, nil
}
