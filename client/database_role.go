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
func (c *client) CreateRole(ctx context.Context, environmentID, instanceID string, create *api.RoleUpsert) (*api.Role, error) {
	payload, err := json.Marshal(create)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/environments/%s/instances/%s/roles", c.HostURL, environmentID, instanceID), strings.NewReader(string(payload)))
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
func (c *client) GetRole(ctx context.Context, environmentID, instanceID, roleName string) (*api.Role, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/environments/%s/instances/%s/roles/%s", c.HostURL, environmentID, instanceID, roleName), nil)
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

// ListRole lists the role in instance.
func (c *client) ListRole(ctx context.Context, environmentID, instanceID string) ([]*api.Role, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/environments/%s/instances/%s/roles", c.HostURL, environmentID, instanceID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var roleList struct {
		Roles []*api.Role `json:"roles"`
	}

	err = json.Unmarshal(body, &roleList)
	if err != nil {
		return nil, err
	}

	return roleList.Roles, nil
}

// UpdateRole updates the role in instance.
func (c *client) UpdateRole(ctx context.Context, environmentID, instanceID, roleName string, patch *api.RoleUpsert) (*api.Role, error) {
	payload, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	paths := []string{}
	if patch.RoleName != roleName {
		paths = append(paths, "role.role_name")
	}
	if patch.Password != nil {
		paths = append(paths, "role.password")
	}
	if patch.ConnectionLimit != nil {
		paths = append(paths, "role.connection_limit")
	}
	if patch.ValidUntil != nil {
		paths = append(paths, "role.valid_until")
	}
	if patch.Attribute != nil {
		paths = append(paths, "role.attribute")
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/environments/%s/instances/%s/roles/%s?update_mask=%s", c.HostURL, environmentID, instanceID, roleName, strings.Join(paths, ",")), strings.NewReader(string(payload)))
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
func (c *client) DeleteRole(ctx context.Context, environmentID, instanceID, roleName string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/environments/%s/instances/%s/roles/%s", c.HostURL, environmentID, instanceID, roleName), nil)
	if err != nil {
		return err
	}

	if _, err := c.doRequest(req); err != nil {
		return err
	}
	return nil
}
