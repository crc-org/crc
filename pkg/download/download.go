package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"

	"github.com/cavaliercoder/grab"
	"github.com/cavaliercoder/grab/grabui"
	"github.com/pkg/errors"
)

// Download function takes sha256sum as hex decoded byte
// something like hex.DecodeString("33daf4c03f86120fdfdc66bddf6bfff4661c7ca11c5d")
func Download(uri, destination string, mode os.FileMode, sha256sum []byte) (string, error) {
	logging.Debugf("Downloading %s to %s", uri, destination)

	client := grab.NewClient()
	client.HTTPClient = &http.Client{Transport: network.HTTPTransport()}
	consoleClient := grabui.NewConsoleClient(client)
	req, err := grab.NewRequest(destination, uri)
	if err != nil {
		return "", errors.Wrapf(err, "unable to get response from %s", uri)
	}
	if sha256sum != nil {
		req.SetChecksum(sha256.New(), sha256sum, true)
	}

	respCh := consoleClient.Do(context.Background(), 3, req)
	resp := <-respCh

	if resp.Err() != nil {
		return "", resp.Err()
	}

	if err := os.Chmod(resp.Filename, mode); err != nil {
		_ = os.Remove(resp.Filename)
		return "", err
	}

	logging.Debugf("Download saved to %v", resp.Filename)
	return resp.Filename, nil
}

type RemoteFile struct {
	uri       string
	sha256sum string
}

func NewRemoteFile(uri, sha256sum string) *RemoteFile {
	return &RemoteFile{
		uri:       uri,
		sha256sum: sha256sum,
	}

}

func (r *RemoteFile) Download(bundlePath string, mode os.FileMode) (string, error) {
	sha256, err := hex.DecodeString(r.sha256sum)
	if err != nil {
		return "", err
	}
	return Download(r.uri, bundlePath, mode, sha256)
}

func (r *RemoteFile) GetSha256Sum() string {
	return r.sha256sum
}
