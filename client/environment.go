package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// CreateEnvironment creates the environment.
func (c *Client) CreateEnvironment(create *api.EnvironmentCreate) (*api.Environment, error) {
	payload, err := json.Marshal(create)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/environment", c.HostURL), strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var env api.Environment
	err = json.Unmarshal(body, &env)
	if err != nil {
		return nil, err
	}

	return &env, nil
}

// GetEnvironment gets the environment by id.
func (c *Client) GetEnvironment(environmentID int) (*api.Environment, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/environment/%d", c.HostURL, environmentID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var env api.Environment
	err = json.Unmarshal(body, &env)
	if err != nil {
		return nil, err
	}

	return &env, nil
}

// ListEnvironment finds all environments.
func (c *Client) ListEnvironment() ([]*api.Environment, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/environment", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	res := []*api.Environment{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// UpdateEnvironment updates the environment.
func (c *Client) UpdateEnvironment(environmentID int, patch *api.EnvironmentPatch) (*api.Environment, error) {
	payload, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/environment/%d", c.HostURL, environmentID), strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var env api.Environment
	err = json.Unmarshal(body, &env)
	if err != nil {
		return nil, err
	}

	return &env, nil
}

// DeleteEnvironment deletes the environment.
func (c *Client) DeleteEnvironment(environmentID int) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/environment/%d", c.HostURL, environmentID), nil)
	if err != nil {
		return err
	}

	if _, err := c.doRequest(req); err != nil {
		return err
	}
	return nil
}
