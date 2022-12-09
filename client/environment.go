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

// CreateEnvironment creates the environment.
func (c *client) CreateEnvironment(ctx context.Context, create *api.EnvironmentUpsert) (*api.Environment, error) {
	payload, err := json.Marshal(create)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/environment", c.HostURL), strings.NewReader(string(payload)))
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
func (c *client) GetEnvironment(ctx context.Context, environmentID int) (*api.Environment, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/environment/%d", c.HostURL, environmentID), nil)
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
func (c *client) ListEnvironment(ctx context.Context, find *api.EnvironmentFind) ([]*api.Environment, error) {
	q, err := query.Values(find)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/environment?%s", c.HostURL, q.Encode()), nil)
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
func (c *client) UpdateEnvironment(ctx context.Context, environmentID int, patch *api.EnvironmentUpsert) (*api.Environment, error) {
	payload, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/environment/%d", c.HostURL, environmentID), strings.NewReader(string(payload)))
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
func (c *client) DeleteEnvironment(ctx context.Context, environmentID int) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/environment/%d", c.HostURL, environmentID), nil)
	if err != nil {
		return err
	}

	if _, err := c.doRequest(req); err != nil {
		return err
	}
	return nil
}
