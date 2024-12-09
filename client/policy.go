package client

import (
	"context"
	"fmt"
	"net/http"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
)

// ListPolicies lists policies in a specific resource.
func (c *client) ListPolicies(ctx context.Context, parent string) (*v1pb.ListPoliciesResponse, error) {
	var url string
	if parent == "" {
		url = fmt.Sprintf("%s/%s/policies", c.url, c.version)
	} else {
		url = fmt.Sprintf("%s/%s/%s/policies", c.url, c.version, parent)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.ListPoliciesResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetPolicy gets a policy in a specific resource.
func (c *client) GetPolicy(ctx context.Context, policyName string) (*v1pb.Policy, error) {
	body, err := c.getResource(ctx, policyName)
	if err != nil {
		return nil, err
	}

	var res v1pb.Policy
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpsertPolicy creates or updates the policy.
func (c *client) UpsertPolicy(ctx context.Context, policy *v1pb.Policy, updateMasks []string) (*v1pb.Policy, error) {
	body, err := c.updateResource(ctx, policy.Name, policy, updateMasks, true /* allow missing = true*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.Policy
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// DeletePolicy deletes the policy.
func (c *client) DeletePolicy(ctx context.Context, policyName string) error {
	return c.deleteResource(ctx, policyName)
}
