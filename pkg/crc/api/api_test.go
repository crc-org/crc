package api

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/code-ready/crc/pkg/crc/machine/fakemachine"

	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApi(t *testing.T) {
	dir, err := ioutil.TempDir("", "api")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	socket := filepath.Join(dir, "api.sock")
	listener, err := net.Listen("unix", socket)
	require.NoError(t, err)
	defer listener.Close()

	client := fakemachine.NewClient()
	api, err := createAPIServerWithListener(listener, client)
	require.NoError(t, err)
	go api.Serve()

	tt := []struct {
		command       string
		clientFailing bool
		expected      map[string]interface{}
	}{
		{
			command: "version",
			expected: map[string]interface{}{
				"CrcVersion":       version.GetCRCVersion(),
				"CommitSha":        version.GetCommitSha(),
				"OpenshiftVersion": version.GetBundleVersion(),
				"Success":          true,
			},
		},
		{
			command: "status",
			expected: map[string]interface{}{
				"Name":             "crc",
				"CrcStatus":        "Running",
				"OpenshiftStatus":  "Running",
				"OpenshiftVersion": "4.5.1",
				"DiskUse":          float64(10000000000),
				"DiskSize":         float64(20000000000),
				"Error":            "",
				"Success":          true,
			},
		},
		{
			command:       "status",
			clientFailing: true,
			expected: map[string]interface{}{
				"Name":             "",
				"CrcStatus":        "",
				"OpenshiftStatus":  "",
				"OpenshiftVersion": "",
				"DiskUse":          float64(0),
				"DiskSize":         float64(0),
				"Error":            "broken",
				"Success":          false,
			},
		},
	}
	for _, test := range tt {
		client.Failing = test.clientFailing
		client, err := net.Dial("unix", socket)
		require.NoError(t, err)

		jsonReq, err := json.Marshal(commandRequest{
			Command: test.command,
			Args:    map[string]string{},
		})
		assert.NoError(t, err)
		_, err = client.Write(jsonReq)
		assert.NoError(t, err)

		payload := make([]byte, 1024)
		n, err := client.Read(payload)
		assert.NoError(t, err)

		var res map[string]interface{}
		assert.NoError(t, json.Unmarshal(payload[0:n], &res))
		assert.Equal(t, test.expected, res)
	}
}
