package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// ListInstance will return instances in environment.
func (c *client) ListInstance(ctx context.Context, showDeleted bool) (*v1pb.ListInstancesResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/instances?showDeleted=%v", c.url, c.version, showDeleted), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.ListInstancesResponse
	err = ProtojsonUnmarshaler.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// GetInstance gets the instance by id.
func (c *client) GetInstance(ctx context.Context, instanceName string) (*v1pb.Instance, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/%s", c.url, c.version, instanceName), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.Instance
	err = ProtojsonUnmarshaler.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// CreateInstance creates the instance.
func (c *client) CreateInstance(ctx context.Context, instanceID string, instance *v1pb.Instance) (*v1pb.Instance, error) {
	payload, err := protojson.Marshal(instance)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/instances?instanceId=%s", c.url, c.version, instanceID), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.Instance
	err = ProtojsonUnmarshaler.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateInstance updates the instance.
func (c *client) UpdateInstance(ctx context.Context, patch *v1pb.Instance, updateMasks []string) (*v1pb.Instance, error) {
	payload, err := protojson.Marshal(patch)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/%s/%s?update_mask=%s", c.url, c.version, patch.Name, strings.Join(updateMasks, ",")), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.Instance
	err = ProtojsonUnmarshaler.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// DeleteInstance deletes the instance.
func (c *client) DeleteInstance(ctx context.Context, instanceName string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/%s/%s", c.url, c.version, instanceName), nil)
	if err != nil {
		return err
	}

	if _, err := c.doRequest(req); err != nil {
		return err
	}
	return nil
}

// UndeleteInstance undeletes the instance.
func (c *client) UndeleteInstance(ctx context.Context, instanceName string) (*v1pb.Instance, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/%s:undelete", c.url, c.version, instanceName), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.Instance
	err = ProtojsonUnmarshaler.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// SyncInstanceSchema will trigger the schema sync for an instance.
func (c *client) SyncInstanceSchema(ctx context.Context, instanceName string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/%s:sync", c.url, c.version, instanceName), nil)

	if err != nil {
		return err
	}

	if _, err := c.doRequest(req); err != nil {
		return err
	}

	return nil
}
