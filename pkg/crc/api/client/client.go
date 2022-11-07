package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client interface {
	Version() (VersionResult, error)
	Status() (ClusterStatusResult, error)
	Start(config StartConfig) (StartResult, error)
	Stop() error
	Delete() error
	WebconsoleURL() (*ConsoleResult, error)
	GetConfig(configs []string) (GetConfigResult, error)
	SetConfig(configs SetConfigRequest) (SetOrUnsetConfigResult, error)
	UnsetConfig(configs []string) (SetOrUnsetConfigResult, error)
	Telemetry(action string) error
	IsPullSecretDefined() (bool, error)
	SetPullSecret(data string) error
}

type HTTPError struct {
	URL        string
	Method     string
	StatusCode int
	Body       string
}

func (err *HTTPError) Error() string {
	return err.Body
}

type client struct {
	client *http.Client
	base   string
}

func New(httpClient *http.Client, baseURL string) Client {
	return &client{
		client: httpClient,
		base:   baseURL,
	}
}

func (c *client) Version() (VersionResult, error) {
	var vr = VersionResult{}
	body, err := c.sendGetRequest("/version")
	if err != nil {
		return vr, err
	}
	err = json.Unmarshal(body, &vr)
	if err != nil {
		return vr, err
	}
	return vr, nil
}

func (c *client) Status() (ClusterStatusResult, error) {
	var sr = ClusterStatusResult{}
	body, err := c.sendGetRequest("/status")
	if err != nil {
		return sr, err
	}
	err = json.Unmarshal(body, &sr)
	if err != nil {
		return sr, err
	}
	return sr, nil
}

func (c *client) Start(config StartConfig) (StartResult, error) {
	var sr = StartResult{}
	var data = new(bytes.Buffer)

	if config != (StartConfig{}) {
		if err := json.NewEncoder(data).Encode(config); err != nil {
			return sr, fmt.Errorf("Failed to encode data to JSON: %w", err)
		}
	}
	body, err := c.sendPostRequest("/start", data)
	if err != nil {
		return sr, err
	}
	err = json.Unmarshal(body, &sr)
	if err != nil {
		return sr, err
	}
	return sr, nil
}

func (c *client) Stop() error {
	_, err := c.sendGetRequest("/stop")
	return err
}

func (c *client) Delete() error {
	_, err := c.sendGetRequest("/delete")
	return err
}

func (c *client) WebconsoleURL() (*ConsoleResult, error) {
	var cr = ConsoleResult{}
	body, err := c.sendGetRequest("/webconsoleurl")
	if err != nil {
		return &cr, err
	}
	err = json.Unmarshal(body, &cr)
	if err != nil {
		return &cr, err
	}
	return &cr, nil
}

func (c *client) GetConfig(configs []string) (GetConfigResult, error) {
	var gcr = GetConfigResult{}
	var escapeConfigs []string
	for _, v := range configs {
		escapeConfigs = append(escapeConfigs, url.QueryEscape(v))
	}
	queryString := strings.Join(escapeConfigs, "&")
	body, err := c.sendGetRequest(fmt.Sprintf("/config?%s", queryString))
	if err != nil {
		return gcr, err
	}
	err = json.Unmarshal(body, &gcr)
	if err != nil {
		return gcr, err
	}
	return gcr, nil
}

func (c *client) SetConfig(configs SetConfigRequest) (SetOrUnsetConfigResult, error) {
	var scr = SetOrUnsetConfigResult{}
	var data = new(bytes.Buffer)

	if len(configs.Properties) == 0 {
		return scr, fmt.Errorf("No config key value pair provided to set")
	}

	if err := json.NewEncoder(data).Encode(configs); err != nil {
		return scr, fmt.Errorf("Failed to encode data to JSON: %w", err)
	}

	body, err := c.sendPostRequest("/config", data)
	if err != nil {
		return scr, err
	}

	err = json.Unmarshal(body, &scr)
	if err != nil {
		return scr, err
	}
	return scr, nil
}

func (c *client) UnsetConfig(configs []string) (SetOrUnsetConfigResult, error) {
	var ucr = SetOrUnsetConfigResult{}
	var data = new(bytes.Buffer)

	cfg := GetOrUnsetConfigRequest{
		Properties: configs,
	}
	if err := json.NewEncoder(data).Encode(cfg); err != nil {
		return ucr, fmt.Errorf("Failed to encode data to JSON: %w", err)
	}
	body, err := c.sendDeleteRequest("/config", data)
	if err != nil {
		return ucr, err
	}
	err = json.Unmarshal(body, &ucr)
	if err != nil {
		return ucr, err
	}
	return ucr, nil
}

func (c *client) Telemetry(action string) error {
	data, err := json.Marshal(TelemetryRequest{
		Action: action,
	})
	if err != nil {
		return fmt.Errorf("Failed to encode data to JSON: %w", err)
	}

	_, err = c.sendPostRequest("/telemetry", bytes.NewReader(data))

	return err
}

func (c *client) IsPullSecretDefined() (bool, error) {
	res, err := c.client.Get(fmt.Sprintf("%s%s", c.base, "/pull-secret"))
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	return res.StatusCode == http.StatusOK, nil
}

func (c *client) SetPullSecret(data string) error {
	_, err := c.sendPostRequest("/pull-secret", bytes.NewReader([]byte(data)))
	if err != nil {
		return err
	}
	return nil
}

func (c *client) sendGetRequest(url string) ([]byte, error) {
	res, err := c.client.Get(fmt.Sprintf("%s%s", c.base, url))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Unknown error reading response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, &HTTPError{
			URL:        url,
			Method:     "GET",
			StatusCode: res.StatusCode,
			Body:       string(body),
		}
	}

	return body, nil
}

func (c *client) sendPostRequest(url string, data io.Reader) ([]byte, error) {
	return c.sendRequest(url, http.MethodPost, data)
}

func (c *client) sendDeleteRequest(url string, data io.Reader) ([]byte, error) {
	return c.sendRequest(url, http.MethodDelete, data)
}

func (c *client) sendRequest(url string, method string, data io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.base, url), data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	switch method {
	case http.MethodPost:
		if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
			return nil, fmt.Errorf("Error occurred sending POST request to : %s : %d", url, res.StatusCode)
		}
	case http.MethodDelete, http.MethodGet:
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Error occurred sending %s request to : %s : %d", method, url, res.StatusCode)
		}
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Unknown error reading response: %w", err)
	}
	return body, nil
}
