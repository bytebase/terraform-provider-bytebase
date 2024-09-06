package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// ListPolicies lists policies in a specific resource.
func (c *client) ListPolicies(ctx context.Context, find *api.PolicyFindMessage) (*api.ListPolicyMessage, error) {
	if find.Type != nil {
		return nil, errors.Errorf("invalid request, list policies cannot specific the policy type")
	}

	var url string
	if find.Parent == "" {
		url = fmt.Sprintf("%s/%s/policies", c.url, c.version)
	} else {
		url = fmt.Sprintf("%s/%s/%s/policies", c.url, c.version, find.Parent)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.ListPolicyMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// GetPolicy gets a policy in a specific resource.
func (c *client) GetPolicy(ctx context.Context, policyName string) (*api.PolicyMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/%s", c.url, c.version, policyName), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.PolicyMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// UpsertPolicy creates or updates the policy.
func (c *client) UpsertPolicy(ctx context.Context, patch *api.PolicyPatchMessage) (*api.PolicyMessage, error) {
	payload, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	paths := []string{}
	if patch.InheritFromParent != nil {
		paths = append(paths, "inherit_from_parent")
	}
	if patch.DeploymentApprovalPolicy != nil ||
		patch.BackupPlanPolicy != nil ||
		patch.SensitiveDataPolicy != nil ||
		patch.AccessControlPolicy != nil {
		paths = append(paths, "payload")
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/%s/%s?allow_missing=true&update_mask=%s", c.url, c.version, patch.Name, strings.Join(paths, ",")), strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.PolicyMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// DeletePolicy deletes the policy.
func (c *client) DeletePolicy(ctx context.Context, policyName string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/%s/%s", c.url, c.version, policyName), nil)
	if err != nil {
		return err
	}

	if _, err := c.doRequest(req); err != nil {
		return err
	}
	return nil
}
