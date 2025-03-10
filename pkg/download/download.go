package download

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/cavaliergopher/grab/v3"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/network/httpproxy"
	"github.com/crc-org/crc/v2/pkg/crc/version"
	"github.com/pkg/errors"
)

// Download function takes sha256sum as hex decoded byte
// something like hex.DecodeString("33daf4c03f86120fdfdc66bddf6bfff4661c7ca11c5d")
func Download(ctx context.Context, uri, destination string, mode os.FileMode, _ []byte) (io.Reader, string, error) {
	logging.Debugf("Downloading %s to %s", uri, destination)

	if ctx == nil {
		panic("ctx is nil, this should not happen")
	}
	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)

	if err != nil {
		return nil, "", errors.Wrapf(err, "unable to get request from %s", uri)
	}
	client := http.Client{Transport: &http.Transport{}}

	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}

	var filename, dir string
	if filepath.Ext(destination) == ".crcbundle" {
		dir = filepath.Dir(destination)
	} else {
		dir = destination
	}
	if disposition, params, _ := mime.ParseMediaType(resp.Header.Get("Content-Disposition")); disposition == "attachment" {
		filename = filepath.Join(dir, params["filename"])
	} else {
		filename = filepath.Join(dir, filepath.Base(resp.Request.URL.Path))
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return nil, "", err
	}

	if err := os.Chmod(filename, mode); err != nil {
		_ = os.Remove(filename)
		return nil, "", err
	}

	return io.TeeReader(resp.Body, file), filename, nil
}

// InMemory takes a URL and returns a ReadCloser object to the downloaded file
// or the file itself if the URL is a file:// URL. In case of failure it returns
// the respective error.
func InMemory(uri string) (io.ReadCloser, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "file" {
		filePath, err := filepath.Abs(u.Path)
		if err != nil {
			return nil, err
		}
		return os.Open(filePath)
	}
	client := grab.NewClient()
	client.HTTPClient = &http.Client{Transport: httpproxy.HTTPTransport()}
	client.UserAgent = version.UserAgent()

	grabReq, err := grab.NewRequest("", uri)
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

func (r *RemoteFile) Download(ctx context.Context, bundlePath string, mode os.FileMode) (io.Reader, string, error) {
	sha256bytes, err := hex.DecodeString(r.sha256sum)
	if err != nil {
		return nil, "", err
	}
	return Download(ctx, r.URI, bundlePath, mode, sha256bytes)
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
