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

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/%s?showDeleted=%v", c.url, c.version, getPolicyRequestName(find), find.ShowDeleted), nil)
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
func (c *client) GetPolicy(ctx context.Context, find *api.PolicyFindMessage) (*api.PolicyMessage, error) {
	if find.Type == nil {
		return nil, errors.Errorf("invalid request, get policy must specific the policy type")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/%s", c.url, c.version, getPolicyRequestName(find)), nil)
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
func (c *client) UpsertPolicy(ctx context.Context, find *api.PolicyFindMessage, patch *api.PolicyPatchMessage) (*api.PolicyMessage, error) {
	if find.Type == nil {
		return nil, errors.Errorf("invalid request, get policy must specific the policy type")
	}

	payload, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	paths := []string{}
	if patch.InheritFromParent != nil {
		paths = append(paths, "policy.inherit_from_parent")
	}
	if patch.DeploymentApprovalPolicy != nil ||
		patch.BackupPlanPolicy != nil ||
		patch.SensitiveDataPolicy != nil ||
		patch.AccessControlPolicy != nil ||
		patch.SQLReviewPolicy != nil {
		paths = append(paths, "policy.payload")
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/%s/%s?allow_missing=true&update_mask=%s", c.url, c.version, getPolicyRequestName(find), strings.Join(paths, ",")), strings.NewReader(string(payload)))
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
func (c *client) DeletePolicy(ctx context.Context, find *api.PolicyFindMessage) error {
	if find.Type == nil {
		return errors.Errorf("invalid request, get policy must specific the policy type")
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/%s/%s", c.url, c.version, getPolicyRequestName(find)), nil)
	if err != nil {
		return err
	}

	if _, err := c.doRequest(req); err != nil {
		return err
	}
	return nil
}

func getPolicyRequestName(find *api.PolicyFindMessage) string {
	paths := []string{}
	if v := find.ProjectID; v != nil {
		paths = append(paths, fmt.Sprintf("projects/%s", *v))
	}
	if v := find.EnvironmentID; v != nil {
		paths = append(paths, fmt.Sprintf("environments/%s", *v))
	}
	if v := find.InstanceID; v != nil {
		paths = append(paths, fmt.Sprintf("instances/%s", *v))
	}
	if v := find.DatabaseName; v != nil {
		paths = append(paths, fmt.Sprintf("databases/%s", *v))
	}

	paths = append(paths, "policies")

	name := strings.Join(paths, "/")
	if v := find.Type; v != nil {
		name = fmt.Sprintf("%s/%s", name, *v)
	}

	return name
}
