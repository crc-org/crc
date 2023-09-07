package download

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/crc-org/crc/pkg/crc/logging"
	"github.com/crc-org/crc/pkg/crc/network/httpproxy"
	"github.com/crc-org/crc/pkg/crc/version"
	"github.com/crc-org/crc/pkg/os/terminal"

	"github.com/cavaliergopher/grab/v3"
	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
)

func doRequest(client *grab.Client, req *grab.Request) (string, error) {
	const minSizeForProgressBar = 100_000_000

	resp := client.Do(req)
	if resp.Size() < minSizeForProgressBar {
		<-resp.Done
		return resp.Filename, resp.Err()
	}

	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	var bar *pb.ProgressBar
	if terminal.IsShowTerminalOutput() {
		bar = pb.Start64(resp.Size())
		bar.Set(pb.Bytes, true)
		// This is the same as the 'Default' template https://github.com/cheggaaa/pb/blob/224e0746e1e7b9c5309d6e2637264bfeb746d043/v3/preset.go#L8-L10
		// except that the 'per second' suffix is changed to '/s' (by default it is ' p/s' which is unexpected)
		progressBarTemplate := `{{with string . "prefix"}}{{.}} {{end}}{{counters . }} {{bar . }} {{percent . }} {{speed . "%s/s" "??/s"}}{{with string . "suffix"}} {{.}}{{end}}`
		bar.SetTemplateString(progressBarTemplate)
		defer bar.Finish()
	}

loop:
	for {
		select {
		case <-t.C:
			if terminal.IsShowTerminalOutput() {
				bar.SetCurrent(resp.BytesComplete())
			}
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
	client.HTTPClient = &http.Client{Transport: httpproxy.HTTPTransport()}
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

// Download takes an URL and the User-Agent string to use in the http request
// it takes  http.RoundTripper as the third argument to be used by the client
func InMemory(url string) (io.ReadCloser, error) {
	client := grab.NewClient()
	client.HTTPClient = &http.Client{Transport: httpproxy.HTTPTransport()}
	client.UserAgent = version.UserAgent()

	grabReq, err := grab.NewRequest("", url)
	if err != nil {
		return nil, fmt.Errorf("failed to create grab request: %w", err)
	}
	// do not write the downloaded file to disk
	grabReq.NoStore = true

	rsp := client.Do(grabReq)
	return rsp.Open()
}

type RemoteFile struct {
	URI       string
	sha256sum string
}

func NewRemoteFile(uri, sha256sum string) *RemoteFile {
	return &RemoteFile{
		URI:       uri,
		sha256sum: sha256sum,
	}

}

func (r *RemoteFile) Download(bundlePath string, mode os.FileMode) (string, error) {
	sha256, err := hex.DecodeString(r.sha256sum)
	if err != nil {
		return "", err
	}
	return Download(r.URI, bundlePath, mode, sha256)
}

func (r *RemoteFile) GetSha256Sum() string {
	return r.sha256sum
}

func (r *RemoteFile) GetSourceFilename() (string, error) {
	u, err := url.Parse(r.URI)
	if err != nil {
		return "", err
	}
	return filepath.Base(u.Path), nil
}
