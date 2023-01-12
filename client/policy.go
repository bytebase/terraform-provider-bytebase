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

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s?showDeleted=%v", c.HostURL, getPolicyRequestName(find), find.ShowDeleted), nil)
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

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s", c.HostURL, getPolicyRequestName(find)), nil)
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
