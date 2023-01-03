package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// ListInstance will return instances in environment.
func (c *client) ListInstance(ctx context.Context, find *api.InstanceFindMessage) (*api.ListInstanceMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/environments/%s/instances?showDeleted=%v", c.HostURL, find.EnvironmentID, find.ShowDeleted), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.ListInstanceMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// GetInstance gets the instance by id.
func (c *client) GetInstance(ctx context.Context, find *api.InstanceFindMessage) (*api.InstanceMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/environments/%s/instances/%s?showDeleted=%v", c.HostURL, find.EnvironmentID, find.InstanceID, find.ShowDeleted), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.InstanceMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// CreateInstance creates the instance.
func (c *client) CreateInstance(ctx context.Context, environmentID, instanceID string, instance *api.InstanceMessage) (*api.InstanceMessage, error) {
	payload, err := json.Marshal(instance)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/environments/%s/instances?instanceId=%s", c.HostURL, environmentID, instanceID), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.InstanceMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateInstance updates the instance.
func (c *client) UpdateInstance(ctx context.Context, environmentID, instanceID string, instance *api.InstanceMessage) (*api.InstanceMessage, error) {
	payload, err := json.Marshal(instance)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/environments/%s/instances/%s", c.HostURL, environmentID, instanceID), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.InstanceMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// DeleteInstance deletes the instance.
func (c *client) DeleteInstance(ctx context.Context, environmentID, instanceID string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/environments/%s/instances/%s", c.HostURL, environmentID, instanceID), nil)
	if err != nil {
		return err
	}

	if _, err := c.doRequest(req); err != nil {
		return err
	}
	return nil
}
