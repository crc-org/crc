package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type Client struct {
	client *http.Client
	base   string
}

func New(client *http.Client, baseURL string) *Client {
	return &Client{
		client: client,
		base:   baseURL,
	}
}

func (c *Client) Version() (VersionResult, error) {
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

func (c *Client) Status() (ClusterStatusResult, error) {
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

func (c *Client) Start(config StartConfig) (StartResult, error) {
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

func (c *Client) Stop() (Result, error) {
	var sr = Result{}
	body, err := c.sendGetRequest("/stop")
	if err != nil {
		return sr, err
	}
	err = json.Unmarshal(body, &sr)
	if err != nil {
		return sr, err
	}
	return sr, nil
}

func (c *Client) Delete() (Result, error) {
	var dr = Result{}
	body, err := c.sendGetRequest("/delete")
	if err != nil {
		return dr, err
	}
	err = json.Unmarshal(body, &dr)
	if err != nil {
		return dr, err
	}
	return dr, nil
}

func (c *Client) WebconsoleURL() (ConsoleResult, error) {
	var cr = ConsoleResult{}
	body, err := c.sendGetRequest("/webconsoleurl")
	if err != nil {
		return cr, err
	}
	err = json.Unmarshal(body, &cr)
	if err != nil {
		return cr, err
	}
	return cr, nil
}

func (c *Client) GetConfig(configs []string) (GetConfigResult, error) {
	var gcr = GetConfigResult{}
	var data = new(bytes.Buffer)

	if len(configs) > 0 {
		cfg := GetOrUnsetConfigRequest{
			Properties: configs,
		}
		if err := json.NewEncoder(data).Encode(cfg); err != nil {
			return gcr, fmt.Errorf("Failed to encode data to JSON: %w", err)
		}
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", c.base, "/config"), data)
	if err != nil {
		return gcr, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return gcr, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return gcr, fmt.Errorf("Unknown error reading response: %w", err)
	}

	err = json.Unmarshal(body, &gcr)
	if err != nil {
		return gcr, err
	}
	return gcr, nil
}

func (c *Client) SetConfig(configs SetConfigRequest) (SetOrUnsetConfigResult, error) {
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

func (c *Client) UnsetConfig(configs []string) (SetOrUnsetConfigResult, error) {
	var ucr = SetOrUnsetConfigResult{}
	var data = new(bytes.Buffer)

	cfg := GetOrUnsetConfigRequest{
		Properties: configs,
	}
	if err := json.NewEncoder(data).Encode(cfg); err != nil {
		return ucr, fmt.Errorf("Failed to encode data to JSON: %w", err)
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s%s", c.base, "/config"), data)
	if err != nil {
		return ucr, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return ucr, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ucr, fmt.Errorf("Unknown error reading response: %w", err)
	}

	err = json.Unmarshal(body, &ucr)
	if err != nil {
		return ucr, err
	}
	return ucr, nil
}

func (c *Client) Telemetry(action string) error {
	data, err := json.Marshal(TelemetryRequest{
		Action: action,
	})
	if err != nil {
		return fmt.Errorf("Failed to encode data to JSON: %w", err)
	}

	body, err := c.sendPostRequest("/telemetry", bytes.NewReader(data))
	if err != nil {
		return err
	}

	var res Result
	if err = json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.Error != "" {
		return errors.New(res.Error)
	}
	return nil
}

func (c *Client) IsPullSecretDefined() (bool, error) {
	res, err := c.client.Get(fmt.Sprintf("%s%s", c.base, "/pull-secret"))
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	return res.StatusCode == http.StatusOK, nil
}

func (c *Client) SetPullSecret(data string) error {
	_, err := c.sendPostRequest("/pull-secret", bytes.NewReader([]byte(data)))
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) sendGetRequest(url string) ([]byte, error) {
	res, err := c.client.Get(fmt.Sprintf("%s%s", c.base, url))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error occurred sending GET request to : %s : %d", url, res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Unknown error reading response: %w", err)
	}
	return body, nil
}

func (c *Client) sendPostRequest(url string, data io.Reader) ([]byte, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", c.base, url), data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Error occurred sending POST request to : %s : %d", url, res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Unknown error reading response: %w", err)
	}
	return body, nil
}
