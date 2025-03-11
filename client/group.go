package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// ListGroup list all groups.
func (c *client) ListGroup(ctx context.Context) (*v1pb.ListGroupsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/groups", c.url, c.version), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.ListGroupsResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// CreateGroup creates the group.
func (c *client) CreateGroup(ctx context.Context, email string, group *v1pb.Group) (*v1pb.Group, error) {
	payload, err := protojson.Marshal(group)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/groups?groupEmail=%s", c.url, c.version, url.QueryEscape(email)), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.Group
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetGroup gets the group by name.
func (c *client) GetGroup(ctx context.Context, name string) (*v1pb.Group, error) {
	body, err := c.getResource(ctx, name)
	if err != nil {
		return nil, err
	}

	var res v1pb.Group
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateGroup updates the group.
func (c *client) UpdateGroup(ctx context.Context, patch *v1pb.Group, updateMasks []string) (*v1pb.Group, error) {
	body, err := c.updateResource(ctx, patch.Name, patch, updateMasks, false /* allow missing = false*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.Group
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// DeleteGroup deletes the group by name.
func (c *client) DeleteGroup(ctx context.Context, name string) error {
	return c.deleteResource(ctx, name)
}
