package segment

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	crcConfig "github.com/crc-org/crc/pkg/crc/config"
	crcErr "github.com/crc-org/crc/pkg/crc/errors"
	"github.com/crc-org/crc/pkg/crc/logging"
	"github.com/crc-org/crc/pkg/crc/telemetry"
	"github.com/crc-org/crc/pkg/crc/version"
	"github.com/stretchr/testify/require"
)

type segmentResponse struct {
	Batch []struct {
		UserID    string `json:"userId"`
		MessageID string `json:"messageId"`
		Traits    struct {
			OS                   string `json:"os"`
			ExperimentalFeatures bool   `json:"enable-experimental-features"`
		} `json:"traits"`
		Properties struct {
			Error     string `json:"error"`
			ErrorType string `json:"error-type"`
			Version   string `json:"version"`
			CPUs      int    `json:"cpus"`
			Remote    bool   `json:"remote"`
		} `json:"properties"`
		Type string `json:"type"`
	} `json:"batch"`
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

func newTestConfig(value string) (*crcConfig.Config, error) {
	storage := crcConfig.NewEmptyInMemoryStorage()
	secretStorage := crcConfig.NewEmptyInMemorySecretStorage()
	config := crcConfig.New(storage, secretStorage)
	crcConfig.RegisterSettings(config)

	if _, err := config.Set(crcConfig.ConsentTelemetry, value); err != nil {
		return nil, err
	}
	if _, err := config.Set(crcConfig.ExperimentalFeatures, true); err != nil {
		return nil, err
	}
	return config, nil
}

func TestClientUploadWithConsentAndWithSerializableError(t *testing.T) {
	body, server := mockServer()
	defer server.Close()
	defer close(body)

	require.NoError(t, os.Setenv("SSH_TTY", "test"))
	defer os.Unsetenv("SSH_TTY")

	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	config, err := newTestConfig("yes")
	require.NoError(t, err)

	uuidFile := filepath.Join(dir, "telemetry")

	c, err := newCustomClient(config, http.DefaultTransport, uuidFile, "", server.URL)
	require.NoError(t, err)

	require.NoError(t, c.UploadCmd(context.Background(), "start", time.Minute, crcErr.ToSerializableError(crcErr.VMNotExist)))
	require.NoError(t, c.Close())

	uuid, err := ioutil.ReadFile(uuidFile)
	require.NoError(t, err)

	select {
	case x := <-body:
		s := segmentResponse{}
		require.NoError(t, json.Unmarshal(x, &s))
		require.Equal(t, s.Batch[0].Type, "identify")
		require.Equal(t, s.Batch[0].UserID, string(uuid))
		require.Equal(t, s.Batch[0].Traits.OS, runtime.GOOS)
		require.Equal(t, s.Batch[0].Traits.ExperimentalFeatures, true)
		require.Equal(t, s.Batch[1].Type, "track")
		require.Equal(t, s.Batch[1].UserID, string(uuid))
		require.Equal(t, s.Batch[1].Properties.Error, crcErr.VMNotExist.Error())
		require.Equal(t, s.Batch[1].Properties.ErrorType, "errors.vmNotExist")
		require.Equal(t, s.Batch[1].Properties.Version, version.GetCRCVersion())
		require.Equal(t, s.Batch[1].Properties.Remote, true)
	default:
		require.Fail(t, "server should receive data")
	}
}

func TestClientUploadWithConsentAndWithoutSerializableError(t *testing.T) {
	body, server := mockServer()
	defer server.Close()
	defer close(body)

	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	config, err := newTestConfig("yes")
	require.NoError(t, err)

	c, err := newCustomClient(config, http.DefaultTransport, filepath.Join(dir, "telemetry"), "", server.URL)
	require.NoError(t, err)

	require.NoError(t, c.UploadCmd(context.Background(), "start", time.Minute, errors.New("an error occurred")))
	require.NoError(t, c.Close())

	select {
	case x := <-body:
		s := segmentResponse{}
		require.NoError(t, json.Unmarshal(x, &s))
		require.Equal(t, s.Batch[0].Type, "identify")
		require.Equal(t, s.Batch[0].Traits.OS, runtime.GOOS)
		require.Equal(t, s.Batch[0].Traits.ExperimentalFeatures, true)
		require.Equal(t, s.Batch[1].Type, "track")
		require.Equal(t, s.Batch[1].Properties.Error, "an error occurred")
		require.Equal(t, s.Batch[1].Properties.ErrorType, "*errors.errorString")
		require.Equal(t, s.Batch[1].Properties.Version, version.GetCRCVersion())
		require.Equal(t, s.Batch[1].Properties.Remote, false)
	default:
		require.Fail(t, "server should receive data")
	}
}

func TestClientUploadWithContext(t *testing.T) {
	body, server := mockServer()
	defer server.Close()
	defer close(body)

	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	config, err := newTestConfig("yes")
	require.NoError(t, err)

	c, err := newCustomClient(config, http.DefaultTransport, filepath.Join(dir, "telemetry"), "", server.URL)
	require.NoError(t, err)

	ctx := telemetry.NewContext(context.Background())
	telemetry.SetCPUs(ctx, 6)
	require.NoError(t, c.UploadCmd(ctx, "start", time.Minute, nil))
	require.NoError(t, c.Close())

	select {
	case x := <-body:
		s := segmentResponse{}
		require.NoError(t, json.Unmarshal(x, &s))
		require.Equal(t, s.Batch[1].Properties.CPUs, 6)
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

	config, err := newTestConfig("no")
	require.NoError(t, err)

	c, err := newCustomClient(config, http.DefaultTransport, filepath.Join(dir, "telemetry"), "", server.URL)
	require.NoError(t, err)

	require.NoError(t, c.UploadCmd(context.Background(), "start", time.Second, errors.New("an error occurred")))
	require.NoError(t, c.Close())

	select {
	case <-body:
		require.Fail(t, "server should not receive data")
	default:
	}
}

func TestClientUploadWithConsentAndCachedIdentify(t *testing.T) {
	body, server := mockServer()
	defer server.Close()
	defer close(body)

	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	config, err := newTestConfig("yes")
	require.NoError(t, err)

	c, err := newCustomClient(config, http.DefaultTransport, filepath.Join(dir, "telemetry"), "", server.URL)
	require.NoError(t, err)

	require.NoError(t, c.UploadCmd(context.Background(), "start", time.Minute, errors.New("an error occurred")))
	require.NoError(t, c.UploadCmd(context.Background(), "stop", time.Minute, errors.New("an error occurred")))
	require.NoError(t, c.Close())

	select {
	case x := <-body:
		s := segmentResponse{}
		require.NoError(t, json.Unmarshal(x, &s))
		require.Equal(t, s.Batch[0].Type, "identify")
		require.Equal(t, s.Batch[1].Type, "track")
		require.Equal(t, s.Batch[2].Type, "track")
	default:
		require.Fail(t, "server should receive data")
	}
}
