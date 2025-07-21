package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// GetRole gets the role by full name.
func (c *client) GetRole(ctx context.Context, name string) (*v1pb.Role, error) {
	body, err := c.getResource(ctx, name, "")
	if err != nil {
		return nil, err
	}

	var res v1pb.Role
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// CreateRole creates the role.
func (c *client) CreateRole(ctx context.Context, roleID string, role *v1pb.Role) (*v1pb.Role, error) {
	payload, err := protojson.Marshal(role)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/roles?roleId=%s", c.url, c.version, roleID), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.Role
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// DeleteRole deletes the role by name.
func (c *client) DeleteRole(ctx context.Context, name string) error {
	return c.deleteResource(ctx, name)
}

// UpdateRole updates the role.
func (c *client) UpdateRole(ctx context.Context, patch *v1pb.Role, updateMasks []string) (*v1pb.Role, error) {
	body, err := c.updateResource(ctx, patch.Name, patch, updateMasks, true /* allow missing = true*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.Role
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ListRole will returns all roles.
func (c *client) ListRole(ctx context.Context) (*v1pb.ListRolesResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/roles", c.url, c.version), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.ListRolesResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
