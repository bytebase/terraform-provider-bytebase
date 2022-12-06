package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-querystring/query"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// ListInstance will return all instances.
func (c *client) ListInstance(ctx context.Context, find *api.InstanceFind) ([]*api.Instance, error) {
	q, err := query.Values(find)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/instance?%s", c.HostURL, q.Encode()), nil)
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
func (c *client) CreateInstance(ctx context.Context, create *api.InstanceCreate) (*api.Instance, error) {
	payload, err := json.Marshal(create)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/instance", c.HostURL), strings.NewReader(string(payload)))
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
func (c *client) GetInstance(ctx context.Context, instanceID int) (*api.Instance, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/instance/%d", c.HostURL, instanceID), nil)
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
func (c *client) UpdateInstance(ctx context.Context, instanceID int, patch *api.InstancePatch) (*api.Instance, error) {
	payload, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/instance/%d", c.HostURL, instanceID), strings.NewReader(string(payload)))
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
func (c *client) DeleteInstance(ctx context.Context, instanceID int) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/instance/%d", c.HostURL, instanceID), nil)
	if err != nil {
		return err
	}

	if _, err := c.doRequest(req); err != nil {
		return err
	}
	return nil
}
