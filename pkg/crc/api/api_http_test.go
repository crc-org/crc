package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"strings"
	"testing"

	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/fakemachine"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockServer struct {
	*server
	client *fakemachine.Client
	config crcConfig.Storage
}

func createDummyPullSecret(t *testing.T) string {
	f, err := ioutil.TempFile("", "kubeconfig")
	assert.NoError(t, err)
	_, err = f.WriteString(constants.OkdPullSecret)
	assert.NoError(t, err)
	err = f.Close()
	assert.NoError(t, err)

	return f.Name()
}

func removePullSecret(t *testing.T, server *mockServer) {
	_, err := server.config.Unset(crcConfig.PullSecretFile)
	assert.NoError(t, err)
}

func newMockServer(pullSecretPath string) *mockServer {
	fakeMachine := fakemachine.NewClient()

	config := setupNewInMemoryConfig()
	_, _ = config.Set(crcConfig.PullSecretFile, pullSecretPath)

	handler := NewHandler(config, fakeMachine, &mockLogger{}, &mockTelemetry{})

	return &mockServer{
		server: newServerWithRoutes(handler),
		client: fakeMachine,
		config: config,
	}
}

func sendRequest(handler http.Handler, request *request) *http.Response {
	url := fmt.Sprintf("/%s", request.resource)
	var data io.Reader
	if request.data != "" {
		data = strings.NewReader(request.data)
	} else {
		data = nil
	}

	req := httptest.NewRequest(request.httpMethod, url, data)
	req.Header.Set("Content-Type", "application/json")
	{
		requestDump, _ := httputil.DumpRequest(req, true)
		fmt.Println(string(requestDump))
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	response := w.Result()
	{
		responseDump, _ := httputil.DumpResponse(response, true)
		fmt.Println(string(responseDump))
	}
	return response
}

type request struct {
	httpMethod string
	resource   string
	data       string
}

type response struct {
	statusCode int
	protoMajor int
	protoMinor int
	// headers
	body string
}

type testCase struct {
	preTestFunc func(t *testing.T, server *mockServer)
	request     request
	failRequest bool
	response    response
}

func get(resource string) request {
	return request{
		httpMethod: http.MethodGet,
		resource:   resource,
	}
}

func post(resource string) request {
	return request{
		httpMethod: http.MethodPost,
		resource:   resource,
	}
}

func delete(resource string) request {
	return request{
		httpMethod: http.MethodDelete,
		resource:   resource,
	}
}

func (req request) String() string {
	return fmt.Sprintf("%s /%s HTTP/1.1", req.httpMethod, req.resource)
}

func (req request) withBody(data string) request {
	req.data = data
	return req
}

func jSon(data string) response {
	return response{
		statusCode: 200,
		protoMajor: 1,
		protoMinor: 1,
		body:       data,
	}
}

func empty() response {
	return response{
		statusCode: 200,
		protoMajor: 1,
		protoMinor: 1,
		body:       `{"Success":true,"Error":""}`,
	}
}

func httpError(statusCode int) response {
	return response{
		statusCode: statusCode,
		protoMajor: 1,
		protoMinor: 1,
	}
}

func (resp response) withBody(body string) response {
	resp.body = body
	return resp
}

var testCases = []testCase{
	// start
	{
		request:  post("start"),
		response: jSon(`{"Success":true,"Status":"","Error":"","ClusterConfig":{"ClusterCACert":"MIIDODCCAiCgAwIBAgIIRVfCKNUa1wIwDQYJ","KubeConfig":"/tmp/kubeconfig","KubeAdminPass":"foobar","ClusterAPI":"https://foo.testing:6443","WebConsoleURL":"https://console.foo.testing:6443","ProxyConfig":null},"KubeletStarted":true}`),
	},
	{
		request:  get("start"),
		response: jSon(`{"Success":true,"Status":"","Error":"","ClusterConfig":{"ClusterCACert":"MIIDODCCAiCgAwIBAgIIRVfCKNUa1wIwDQYJ","KubeConfig":"/tmp/kubeconfig","KubeAdminPass":"foobar","ClusterAPI":"https://foo.testing:6443","WebConsoleURL":"https://console.foo.testing:6443","ProxyConfig":null},"KubeletStarted":true}`),
	},

	// start with failure
	{
		request:     post("start"),
		failRequest: true,
		response:    httpError(500).withBody("Failed to start\n"),
	},
	{
		request:     get("start"),
		failRequest: true,
		response:    httpError(500).withBody("Failed to start\n"),
	},

	// stop
	{
		request:  post("stop"),
		response: empty(),
	},
	{
		request:  get("stop"),
		response: empty(),
	},

	// stop with failure
	{
		request:     post("stop"),
		failRequest: true,
		// error message comes from fakemachine
		response: httpError(500).withBody("stop failed\n"),
	},
	{
		request:     get("stop"),
		failRequest: true,
		// error message comes from fakemachine
		response: httpError(500).withBody("stop failed\n"),
	},

	// poweroff
	{
		request:  post("poweroff"),
		response: empty(),
	},

	// poweroff with failure
	{
		request:     post("poweroff"),
		failRequest: true,
		// error message comes from fakemachine
		response: httpError(500).withBody("poweroff failed\n"),
	},

	// status
	{
		request:  get("status"),
		response: jSon(`{"CrcStatus":"Running","OpenshiftStatus":"Running","OpenshiftVersion":"4.5.1","DiskUse":10000000000,"DiskSize":20000000000,"Error":"","Success":true}`),
	},

	// status with failure
	{
		request:     get("status"),
		failRequest: true,
		// error message comes from fakemachine
		response: httpError(500).withBody("broken\n"),
	},

	// delete
	{
		request:  delete("delete"),
		response: empty(),
	},
	{
		request:  get("delete"),
		response: empty(),
	},

	// delete with failure
	{
		request:     delete("delete"),
		failRequest: true,
		// error message comes from fakemachine
		response: httpError(500).withBody("delete failed\n"),
	},
	{
		request:     get("delete"),
		failRequest: true,
		// error message comes from fakemachine
		response: httpError(500).withBody("delete failed\n"),
	},

	// version
	{
		request:  get("version"),
		response: jSon(fmt.Sprintf(`{"CrcVersion":"%s","CommitSha":"%s","OpenshiftVersion":"%s","Success":true}`, version.GetCRCVersion(), version.GetCommitSha(), version.GetBundleVersion())),
	},

	// version never fails

	// webconsoleurl
	{
		request:  get("webconsoleurl"),
		response: jSon(`{"ClusterConfig":{"ClusterCACert":"MIIDODCCAiCgAwIBAgIIRVfCKNUa1wIwDQYJ","KubeConfig":"/tmp/kubeconfig","KubeAdminPass":"foobar","ClusterAPI":"https://foo.testing:6443","WebConsoleURL":"https://console.foo.testing:6443","ProxyConfig":null},"Success":true,"Error":""}`),
	},

	// webconsoleurl with failure
	{
		request:     get("webconsoleurl"),
		failRequest: true,
		// error message comes from fakemachine
		response: httpError(500).withBody("console failed\n"),
	},

	// config
	{
		request:  get("config?cpus"),
		response: jSon(`{"Success":true,"Error":"","Configs":{"cpus":4}}`),
	},
	{
		request:  post("config?cpus").withBody("xx"),
		response: httpError(500).withBody("invalid character 'x' looking for beginning of value\n"),
	},
	{
		request:  delete("config?cpus"),
		response: httpError(500).withBody("unexpected end of JSON input\n"),
	},
	{
		request:  get("config?cpus").withBody("xx"),
		response: jSon(`{"Success":true,"Error":"","Configs":{"cpus":4}}`),
	},

	// logs
	{
		request:  get("logs"),
		response: jSon(`{"Success":true,"Messages":["message 1","message 2","message 3"]}`),
	},

	// logs never fails

	// telemetry
	{
		request:  get("telemetry"),
		response: httpError(500).withBody("unexpected end of JSON input\n"),
	},
	{
		request:  post("telemetry"),
		response: httpError(500).withBody("unexpected end of JSON input\n"),
	},

	// telemetry with failure
	{
		request:     get("telemetry"),
		failRequest: true,
		response:    httpError(500).withBody("unexpected end of JSON input\n"),
	},
	{
		request:     post("telemetry"),
		failRequest: true,
		response:    httpError(500).withBody("unexpected end of JSON input\n"),
	},

	// pull-secret
	{
		request: get("pull-secret"),
		// other 404 return "not found", and others "404 not found"
		response: httpError(404),
	},
	{
		request:  post("pull-secret"),
		response: httpError(500).withBody("empty pull secret\n"),
	},

	// pull-secret with failure
	{
		preTestFunc: removePullSecret,
		request:     get("pull-secret"),
		failRequest: true,
		// other 404 return "not found", and others "404 not found"
		response: httpError(404),
	},
	{
		request:     post("pull-secret"),
		failRequest: true,
		response:    httpError(500).withBody("empty pull secret\n"),
	},

	// not found
	{
		request:  get("notfound"),
		response: httpError(404).withBody("Not Found\n"),
	},

	// config
	{
		request:  get("config?cpus"),
		response: jSon(`{"Success":true,"Error":"","Configs":{"cpus":4}}`),
	},
}

var invalidHTTPMethods = []testCase{
	// start
	{
		request:  delete("start"),
		response: httpError(404).withBody("Not Found\n"),
	},

	// stop
	{
		request:  delete("stop"),
		response: httpError(404).withBody("Not Found\n"),
	},

	// poweroff
	{
		request:  get("poweroff"),
		response: httpError(404).withBody("Not Found\n"),
	},
	{
		request:  delete("poweroff"),
		response: httpError(404).withBody("Not Found\n"),
	},

	// status
	{
		request:  post("status"),
		response: httpError(404).withBody("Not Found\n"),
	},
	{
		request:  delete("status"),
		response: httpError(404).withBody("Not Found\n"),
	},

	// delete
	{
		request:  post("delete"),
		response: httpError(404).withBody("Not Found\n"),
	},

	// version
	{
		request:  post("version"),
		response: httpError(404).withBody("Not Found\n"),
	},
	{
		request:  delete("version"),
		response: httpError(404).withBody("Not Found\n"),
	},

	// webconsoleurl
	{
		request:  post("webconsoleurl"),
		response: httpError(404).withBody("Not Found\n"),
	},
	{
		request:  delete("webconsoleurl"),
		response: httpError(404).withBody("Not Found\n"),
	},

	// logs
	{
		request:  post("logs"),
		response: httpError(404).withBody("Not Found\n"),
	},
	{
		request:  delete("logs"),
		response: httpError(404).withBody("Not Found\n"),
	},

	// telemetry
	{
		request:  delete("telemetry"),
		response: httpError(404).withBody("Not Found\n"),
	},

	// pull-secret
	{
		request: delete("pull-secret"),
		// other 404 return "not found", and others "404 not found"
		response: httpError(404).withBody("Not Found\n"),
	},
}

func testOne(t *testing.T, testCase *testCase, server *mockServer) {
	server.client.Failing = testCase.failRequest
	if testCase.preTestFunc != nil {
		testCase.preTestFunc(t, server)
	}
	resp := sendRequest(server.Handler(), &testCase.request)

	require.Equal(t, testCase.response.statusCode, resp.StatusCode, testCase.request)
	require.Equal(t, testCase.response.protoMajor, resp.ProtoMajor, testCase.request)
	require.Equal(t, testCase.response.protoMinor, resp.ProtoMinor, testCase.request)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err, testCase.request)
	require.Equal(t, testCase.response.body, string(body), testCase.request)
	fmt.Println("-----")
}

func TestRequests(t *testing.T) {
	pullSecretPath := createDummyPullSecret(t)
	defer os.Remove(pullSecretPath)
	server := newMockServer(pullSecretPath)

	for i := range testCases {
		testOne(t, &testCases[i], server)
	}

	for i := range invalidHTTPMethods {
		testOne(t, &testCases[i], server)
	}
}

func TestRoutes(t *testing.T) {
	// this checks that we have test cases for all routes registered with the `api` entrypoint

	var routes = map[string][]string{}
	for _, testCase := range testCases {
		// Add leading '/', remove trailing '?....'
		pattern := fmt.Sprintf("/%s", strings.SplitN(testCase.request.resource, "?", 2)[0])
		if _, ok := routes[pattern]; !ok {
			routes[pattern] = []string{}
		}
		routes[pattern] = append(routes[pattern], testCase.request.httpMethod)
	}

	server := newMockServer("")
	for pattern, methodMap := range server.routes {
		assert.Contains(t, routes, pattern)
		for method := range methodMap {
			assert.Contains(t, routes[pattern], method, "routes[%s][%s] is missing from the API testcases", pattern, method)
		}
	}
}
