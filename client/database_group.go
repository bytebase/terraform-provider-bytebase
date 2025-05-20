package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// ListDatabaseGroup list all database groups in a project.
func (c *client) ListDatabaseGroup(ctx context.Context, project string) (*v1pb.ListDatabaseGroupsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/%s/databaseGroups", c.url, c.version, project), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.ListDatabaseGroupsResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// CreateDatabaseGroup creates the database group.
func (c *client) CreateDatabaseGroup(ctx context.Context, project, groupID string, group *v1pb.DatabaseGroup) (*v1pb.DatabaseGroup, error) {
	payload, err := protojson.Marshal(group)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/%s/databaseGroups?databaseGroupId=%s", c.url, c.version, project, groupID), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.DatabaseGroup
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetDatabaseGroup gets the database group by name.
func (c *client) GetDatabaseGroup(ctx context.Context, name string, view v1pb.DatabaseGroupView) (*v1pb.DatabaseGroup, error) {
	// TODO(ed): query
	body, err := c.getResource(ctx, name, fmt.Sprintf("view=%s", view.String()))
	if err != nil {
		return nil, err
	}

	var res v1pb.DatabaseGroup
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateDatabaseGroup updates the database group.
func (c *client) UpdateDatabaseGroup(ctx context.Context, patch *v1pb.DatabaseGroup, updateMasks []string) (*v1pb.DatabaseGroup, error) {
	body, err := c.updateResource(ctx, patch.Name, patch, updateMasks, false /* allow missing = false*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.DatabaseGroup
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// DeleteDatabaseGroup deletes the database group by name.
func (c *client) DeleteDatabaseGroup(ctx context.Context, name string) error {
	return c.deleteResource(ctx, name)
}
