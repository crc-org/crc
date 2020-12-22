package segment

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/segmentio/analytics-go"
	"github.com/stretchr/testify/require"
)

type segmentResponse struct {
	Batch []struct {
		AnonymousID string `json:"anonymousId"`
		MessageID   string `json:"messageId"`
		Traits      struct {
			Error string `json:"error"`
		} `json:"traits"`
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
		buf := bytes.NewBuffer(nil)
		io.Copy(buf, r.Body) // nolint

		var v interface{}
		err := json.Unmarshal(buf.Bytes(), &v)
		if err != nil {
			panic(err)
		}

		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			panic(err)
		}

		done <- b
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

	client, err := analytics.NewWithConfig("dummykey", analytics.Config{
		DefaultContext: &analytics.Context{
			App: analytics.AppInfo{
				Name:    "crc",
				Version: "1.20.0",
			},
		},
		Endpoint: server.URL,
	})
	require.NoError(t, err)

	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	config, err := newTestConfig(true)
	require.NoError(t, err)

	c := &Client{segmentClient: client, config: config, telemetryFilePath: filepath.Join(dir, "telemetry")}

	require.NoError(t, c.Upload(errors.New("an error occurred")))

	s := segmentResponse{}
	select {
	case x, ok := <-body:
		if ok {
			err = json.Unmarshal(x, &s)
			require.NoError(t, err)
		}
	default:
	}
	require.Equal(t, s.Batch[0].Traits.Error, "an error occurred")
	require.Equal(t, s.Context.App.Name, "crc")
	require.Equal(t, s.Context.App.Version, "1.20.0")
}

func TestClientUploadWithOutConsent(t *testing.T) {
	body, server := mockServer()
	defer server.Close()
	defer close(body)

	client, err := analytics.NewWithConfig("dummykey", analytics.Config{
		DefaultContext: &analytics.Context{
			App: analytics.AppInfo{
				Name:    "crc",
				Version: "1.20.0",
			},
		},
		Endpoint: server.URL,
	})
	require.NoError(t, err)

	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	config, err := newTestConfig(false)
	require.NoError(t, err)

	c := &Client{segmentClient: client, config: config, telemetryFilePath: filepath.Join(dir, "telemetry")}

	require.NoError(t, c.Upload(errors.New("an error occurred")))

	s := segmentResponse{}
	select {
	case x, ok := <-body:
		if ok {
			err = json.Unmarshal(x, &s)
			require.NoError(t, err)
		}
	default:
	}

	require.Len(t, s.Batch, 0)
}
