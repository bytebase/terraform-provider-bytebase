package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// CreateEnvironment creates the environment.
func (c *client) CreateEnvironment(ctx context.Context, environmentID string, create *api.EnvironmentMessage) (*api.EnvironmentMessage, error) {
	payload, err := json.Marshal(create)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/environments?environmentId=%s", c.url, c.version, environmentID), strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var env api.EnvironmentMessage
	err = json.Unmarshal(body, &env)
	if err != nil {
		return nil, err
	}

	return &env, nil
}

// GetEnvironment gets the environment by id.
func (c *client) GetEnvironment(ctx context.Context, environmentName string) (*api.EnvironmentMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/%s", c.url, c.version, environmentName), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var env api.EnvironmentMessage
	err = json.Unmarshal(body, &env)
	if err != nil {
		return nil, err
	}

	return &env, nil
}

// ListEnvironment finds all environments.
func (c *client) ListEnvironment(ctx context.Context, showDeleted bool) (*api.ListEnvironmentMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/environments?showDeleted=%v", c.url, c.version, showDeleted), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.ListEnvironmentMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateEnvironment updates the environment.
func (c *client) UpdateEnvironment(ctx context.Context, patch *api.EnvironmentPatchMessage) (*api.EnvironmentMessage, error) {
	payload, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	paths := []string{}
	if patch.Title != nil {
		paths = append(paths, "title")
	}
	if patch.Order != nil {
		paths = append(paths, "order")
	}
	if patch.Tier != nil {
		paths = append(paths, "tier")
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/%s/%s?update_mask=%s", c.url, c.version, patch.Name, strings.Join(paths, ",")), strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var env api.EnvironmentMessage
	err = json.Unmarshal(body, &env)
	if err != nil {
		return nil, err
	}

	return &env, nil
}

// DeleteEnvironment deletes the environment.
func (c *client) DeleteEnvironment(ctx context.Context, environmentName string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/%s/%s", c.url, c.version, environmentName), nil)
	if err != nil {
		return err
	}

	if _, err := c.doRequest(req); err != nil {
		return err
	}
	return nil
}

// UndeleteEnvironment undeletes the environment.
func (c *client) UndeleteEnvironment(ctx context.Context, environmentName string) (*api.EnvironmentMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/%s:undelete", c.url, c.version, environmentName), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.EnvironmentMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
