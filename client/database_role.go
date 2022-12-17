package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// CreateRole creates the role in the instance.
func (c *client) CreateRole(ctx context.Context, instanceID int, create *api.RoleUpsert) (*api.Role, error) {
	payload, err := json.Marshal(create)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/instance/%d/role", c.HostURL, instanceID), strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var role api.Role
	err = json.Unmarshal(body, &role)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// GetRole gets the role by instance id and role name.
func (c *client) GetRole(ctx context.Context, instanceID int, roleName string) (*api.Role, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/instance/%d/role/%s", c.HostURL, instanceID, roleName), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var role api.Role
	err = json.Unmarshal(body, &role)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// UpdateRole updates the role in instance.
func (c *client) UpdateRole(ctx context.Context, instanceID int, roleName string, patch *api.RoleUpsert) (*api.Role, error) {
	payload, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/instance/%d/role/%s", c.HostURL, instanceID, roleName), strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var role api.Role
	err = json.Unmarshal(body, &role)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// DeleteRole deletes the role in the instance.
func (c *client) DeleteRole(ctx context.Context, instanceID int, roleName string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/instance/%d/role/%s", c.HostURL, instanceID, roleName), nil)
	if err != nil {
		return err
	}

	if _, err := c.doRequest(req); err != nil {
		return err
	}
	return nil
}
