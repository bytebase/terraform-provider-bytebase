package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// GetProject gets the project by project full name.
func (c *client) GetProject(ctx context.Context, projectName string) (*v1pb.Project, error) {
	body, err := c.getResource(ctx, projectName)
	if err != nil {
		return nil, err
	}

	var res v1pb.Project
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetProjectIAMPolicy gets the project IAM policy by project full name.
func (c *client) GetProjectIAMPolicy(ctx context.Context, projectName string) (*v1pb.IamPolicy, error) {
	body, err := c.getResource(ctx, fmt.Sprintf("%s:getIamPolicy", projectName))
	if err != nil {
		return nil, err
	}

	var res v1pb.IamPolicy
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetProjectIAMPolicy sets the project IAM policy.
func (c *client) SetProjectIAMPolicy(ctx context.Context, projectName string, update *v1pb.SetIamPolicyRequest) (*v1pb.IamPolicy, error) {
	payload, err := protojson.Marshal(update)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/%s:setIamPolicy", c.url, c.version, projectName), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.IamPolicy
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ListProject list the projects.
func (c *client) ListProject(ctx context.Context, showDeleted bool) (*v1pb.ListProjectsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/projects?showDeleted=%v", c.url, c.version, showDeleted), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.ListProjectsResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// CreateProject creates the project.
func (c *client) CreateProject(ctx context.Context, projectID string, project *v1pb.Project) (*v1pb.Project, error) {
	payload, err := protojson.Marshal(project)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/projects?projectId=%s", c.url, c.version, projectID), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.Project
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateProject updates the project.
func (c *client) UpdateProject(ctx context.Context, patch *v1pb.Project, updateMasks []string) (*v1pb.Project, error) {
	body, err := c.updateResource(ctx, patch.Name, patch, updateMasks, false /* allow missing = false*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.Project
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// DeleteProject deletes the project.
func (c *client) DeleteProject(ctx context.Context, projectName string) error {
	return c.deleteResource(ctx, projectName)
}

// UndeleteProject undeletes the project.
func (c *client) UndeleteProject(ctx context.Context, projectName string) (*v1pb.Project, error) {
	body, err := c.undeleteResource(ctx, projectName)
	if err != nil {
		return nil, err
	}

	var res v1pb.Project
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
