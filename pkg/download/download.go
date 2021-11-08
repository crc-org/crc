package download

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"

	"github.com/cavaliercoder/grab"
	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
)

func doRequest(client *grab.Client, req *grab.Request) (string, error) {
	resp := client.Do(req)

	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	bar := pb.Start64(resp.Size())
	bar.Set(pb.Bytes, true)
	defer bar.Finish()

loop:
	for {
		select {
		case <-t.C:
			bar.SetCurrent(resp.BytesComplete())
		case <-resp.Done:
			break loop
		}
	}

	return resp.Filename, resp.Err()
}

// Download function takes sha256sum as hex decoded byte
// something like hex.DecodeString("33daf4c03f86120fdfdc66bddf6bfff4661c7ca11c5d")
func Download(uri, destination string, mode os.FileMode, sha256sum []byte) (string, error) {
	logging.Debugf("Downloading %s to %s", uri, destination)

	client := grab.NewClient()
	client.HTTPClient = &http.Client{Transport: network.HTTPTransport()}
	req, err := grab.NewRequest(destination, uri)
	if err != nil {
		return "", errors.Wrapf(err, "unable to get request from %s", uri)
	}
	if sha256sum != nil {
		req.SetChecksum(sha256.New(), sha256sum, true)
	}

	filename, err := doRequest(client, req)
	if err != nil {
		return "", err
	}

	if err := os.Chmod(filename, mode); err != nil {
		_ = os.Remove(filename)
		return "", err
	}

	logging.Debugf("Download saved to %v", filename)
	return filename, nil
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
