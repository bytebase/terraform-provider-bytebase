package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// CreateEnvironment creates the environment.
func (c *client) CreateEnvironment(ctx context.Context, environmentID string, create *v1pb.Environment) (*v1pb.Environment, error) {
	payload, err := protojson.Marshal(create)
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

	var env v1pb.Environment
	if err := ProtojsonUnmarshaler.Unmarshal(body, &env); err != nil {
		return nil, err
	}

	return &env, nil
}

// GetEnvironment gets the environment by full name.
func (c *client) GetEnvironment(ctx context.Context, environmentName string) (*v1pb.Environment, error) {
	body, err := c.getResource(ctx, environmentName)
	if err != nil {
		return nil, err
	}

	var env v1pb.Environment
	if err := ProtojsonUnmarshaler.Unmarshal(body, &env); err != nil {
		return nil, err
	}

	return &env, nil
}

// ListEnvironment finds all environments.
func (c *client) ListEnvironment(ctx context.Context, showDeleted bool) (*v1pb.ListEnvironmentsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/environments?showDeleted=%v", c.url, c.version, showDeleted), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.ListEnvironmentsResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateEnvironment updates the environment.
func (c *client) UpdateEnvironment(ctx context.Context, patch *v1pb.Environment, updateMasks []string) (*v1pb.Environment, error) {
	body, err := c.updateResource(ctx, patch.Name, patch, updateMasks, false /* allow missing = false*/)
	if err != nil {
		return nil, err
	}

	var env v1pb.Environment
	if err := ProtojsonUnmarshaler.Unmarshal(body, &env); err != nil {
		return nil, err
	}

	return &env, nil
}

// DeleteEnvironment deletes the environment.
func (c *client) DeleteEnvironment(ctx context.Context, environmentName string) error {
	return c.deleteResource(ctx, environmentName)
}

// UndeleteEnvironment undeletes the environment.
func (c *client) UndeleteEnvironment(ctx context.Context, environmentName string) (*v1pb.Environment, error) {
	body, err := c.undeleteResource(ctx, environmentName)
	if err != nil {
		return nil, err
	}

	var res v1pb.Environment
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
