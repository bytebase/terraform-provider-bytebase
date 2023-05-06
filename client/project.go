package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// GetProject gets the project by resource id.
func (c *client) GetProject(ctx context.Context, projectID string, showDeleted bool) (*api.ProjectMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/projects/%s?showDeleted=%v", c.url, c.version, projectID, showDeleted), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.ProjectMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// ListProject list the projects.
func (c *client) ListProject(ctx context.Context, showDeleted bool) (*api.ListProjectMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/projects?showDeleted=%v", c.url, c.version, showDeleted), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.ListProjectMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// CreateProject creates the project.
func (c *client) CreateProject(ctx context.Context, project *api.ProjectMessage) (*api.ProjectMessage, error) {
	payload, err := json.Marshal(project)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/projects", c.url, c.version), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.ProjectMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateProject updates the project.
func (c *client) UpdateProject(ctx context.Context, projectID string, patch *api.ProjectPatchMessage) (*api.ProjectMessage, error) {
	payload, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	paths := []string{}
	if patch.Title != nil {
		paths = append(paths, "title")
	}
	if patch.Key != nil {
		paths = append(paths, "key")
	}
	if patch.Workflow != nil {
		paths = append(paths, "workflow")
	}
	if patch.TenantMode != nil {
		paths = append(paths, "tenant_mode")
	}
	if patch.DBNameTemplate != nil {
		paths = append(paths, "db_name_template")
	}
	if patch.SchemaChange != nil {
		paths = append(paths, "schema_change")
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/%s/projects/%s?update_mask=%s", c.url, c.version, projectID, strings.Join(paths, ",")), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.ProjectMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// DeleteProject deletes the project.
func (c *client) DeleteProject(ctx context.Context, projectID string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/%s/projects/%s", c.url, c.version, projectID), nil)
	if err != nil {
		return err
	}

	if _, err := c.doRequest(req); err != nil {
		return err
	}
	return nil
}

// UndeleteProject undeletes the project.
func (c *client) UndeleteProject(ctx context.Context, projectID string) (*api.InstanceMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/projects/%s:undelete", c.url, c.version, projectID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.InstanceMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
