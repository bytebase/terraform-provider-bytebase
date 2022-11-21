package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// ListInstance will return all instances.
func (c *Client) ListInstance() ([]*api.Instance, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/instance", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	orders := []*api.Instance{}
	err = json.Unmarshal(body, &orders)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// CreateInstance creates the instance.
func (c *Client) CreateInstance(create *api.InstanceCreate) (*api.Instance, error) {
	payload, err := json.Marshal(create)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/instance", c.HostURL), strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var instance api.Instance
	err = json.Unmarshal(body, &instance)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

// GetInstance gets the instance by id.
func (c *Client) GetInstance(instanceID int) (*api.Instance, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/instance/%d", c.HostURL, instanceID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var instance api.Instance
	err = json.Unmarshal(body, &instance)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

// UpdateInstance updates the instance.
func (c *Client) UpdateInstance(instanceID int, patch *api.InstancePatch) (*api.Instance, error) {
	payload, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/instance/%d", c.HostURL, instanceID), strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var instance api.Instance
	err = json.Unmarshal(body, &instance)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

// DeleteInstance deletes the instance.
func (c *Client) DeleteInstance(instanceID int) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/instance/%d", c.HostURL, instanceID), nil)
	if err != nil {
		return err
	}

	if _, err := c.doRequest(req); err != nil {
		return err
	}
	return nil
}
