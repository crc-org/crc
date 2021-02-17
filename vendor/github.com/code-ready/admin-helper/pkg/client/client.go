package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/code-ready/admin-helper/pkg/types"
)

type Client struct {
	client *http.Client
	base   string
}

func New(client *http.Client, base string) *Client {
	return &Client{
		client: client,
		base:   base,
	}
}

func (c *Client) Version() (string, error) {
	res, err := c.client.Get(fmt.Sprintf("%s%s", c.base, "/version"))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

func (c *Client) Add(req *types.AddRequest) error {
	bin, err := json.Marshal(req)
	if err != nil {
		return err
	}
	res, err := c.client.Post(fmt.Sprintf("%s%s", c.base, "/add"), "application/json", bytes.NewReader(bin))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		err, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			return fmt.Errorf("error while reading error message: %v", readErr)
		}
		return errors.New(strings.TrimSpace(string(err)))
	}
	return nil
}

func (c *Client) Remove(req *types.RemoveRequest) error {
	bin, err := json.Marshal(req)
	if err != nil {
		return err
	}
	res, err := c.client.Post(fmt.Sprintf("%s%s", c.base, "/remove"), "application/json", bytes.NewReader(bin))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		err, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			return fmt.Errorf("error while reading error message: %v", readErr)
		}
		return errors.New(strings.TrimSpace(string(err)))
	}
	return nil
}

func (c *Client) Clean(req *types.CleanRequest) error {
	bin, err := json.Marshal(req)
	if err != nil {
		return err
	}
	res, err := c.client.Post(fmt.Sprintf("%s%s", c.base, "/clean"), "application/json", bytes.NewReader(bin))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		err, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			return fmt.Errorf("error while reading error message: %v", readErr)
		}
		return errors.New(strings.TrimSpace(string(err)))
	}
	return nil
}
