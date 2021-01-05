package segment

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/stretchr/testify/require"
)

type segmentResponse struct {
	Batch []struct {
		AnonymousID string `json:"anonymousId"`
		MessageID   string `json:"messageId"`
		Traits      struct {
			OS string `json:"os"`
		} `json:"traits"`
		Properties struct {
			Error string `json:"error"`
		} `json:"properties"`
		Type string `json:"type"`
	} `json:"batch"`
	Context struct {
		App struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"app"`
		Library struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"library"`
	} `json:"context"`
	MessageID string `json:"messageId"`
}

func mockServer() (chan []byte, *httptest.Server) {
	done := make(chan []byte, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		bin, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logging.Error(err)
			return
		}
		done <- bin
	}))

	return done, server
}

func newTestConfig(value bool) (*crcConfig.Config, error) {
	storage := crcConfig.NewEmptyInMemoryStorage()
	config := crcConfig.New(storage)
	cmdConfig.RegisterSettings(config)

	if _, err := config.Set(cmdConfig.ConsentTelemetry, value); err != nil {
		return nil, err
	}
	return config, nil
}

func TestClientUploadWithConsent(t *testing.T) {
	body, server := mockServer()
	defer server.Close()
	defer close(body)

	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	config, err := newTestConfig(true)
	require.NoError(t, err)

	c, err := newCustomClient(config, filepath.Join(dir, "telemetry"), server.URL)
	require.NoError(t, err)

	require.NoError(t, c.Upload("start", errors.New("an error occurred")))
	require.NoError(t, c.Close())

	select {
	case x := <-body:
		s := segmentResponse{}
		require.NoError(t, json.Unmarshal(x, &s))
		require.Equal(t, s.Batch[0].Type, "identify")
		require.Equal(t, s.Batch[0].Traits.OS, runtime.GOOS)
		require.Equal(t, s.Batch[1].Type, "track")
		require.Equal(t, s.Batch[1].Properties.Error, "an error occurred")
		require.Equal(t, s.Context.App.Name, "crc")
		require.Equal(t, s.Context.App.Version, version.GetCRCVersion())
	default:
		require.Fail(t, "server should receive data")
	}
}

func TestClientUploadWithOutConsent(t *testing.T) {
	body, server := mockServer()
	defer server.Close()
	defer close(body)

	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	config, err := newTestConfig(false)
	require.NoError(t, err)

	c, err := newCustomClient(config, filepath.Join(dir, "telemetry"), server.URL)
	require.NoError(t, err)

	require.NoError(t, c.Upload("start", errors.New("an error occurred")))
	require.NoError(t, c.Close())

	select {
	case <-body:
		require.Fail(t, "server should not receive data")
	default:
	}
}
