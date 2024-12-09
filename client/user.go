package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// ListUser list all users.
func (c *client) ListUser(ctx context.Context, showDeleted bool) (*v1pb.ListUsersResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/users?showDeleted=%v", c.url, c.version, showDeleted), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.ListUsersResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// CreateUser creates the user.
func (c *client) CreateUser(ctx context.Context, user *v1pb.User) (*v1pb.User, error) {
	payload, err := protojson.Marshal(user)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/users", c.url, c.version), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.User
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetUser gets the user by name.
func (c *client) GetUser(ctx context.Context, userName string) (*v1pb.User, error) {
	body, err := c.getResource(ctx, userName)
	if err != nil {
		return nil, err
	}

	var res v1pb.User
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateUser updates the user.
func (c *client) UpdateUser(ctx context.Context, patch *v1pb.User, updateMasks []string) (*v1pb.User, error) {
	body, err := c.updateResource(ctx, patch.Name, patch, updateMasks, false /* allow missing = false*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.User
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// DeleteUser deletes the user by name.
func (c *client) DeleteUser(ctx context.Context, userName string) error {
	return c.deleteResource(ctx, userName)
}

// UndeleteUser undeletes the user by name.
func (c *client) UndeleteUser(ctx context.Context, userName string) (*v1pb.User, error) {
	body, err := c.undeleteResource(ctx, userName)
	if err != nil {
		return nil, err
	}

	var res v1pb.User
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
