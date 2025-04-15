package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/bytebase/terraform-provider-bytebase/api"
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

func buildProjectQuery(filter *api.ProjectFilter) string {
	params := []string{}
	showDeleted := v1pb.State_DELETED == filter.State

	if v := filter.Query; v != "" {
		params = append(params, fmt.Sprintf(`(name.matches("%s") || resource_id.matches("%s"))`, strings.ToLower(v), strings.ToLower(v)))
	}
	if filter.ExcludeDefault {
		params = append(params, "exclude_default == true")
	}
	if showDeleted {
		params = append(params, fmt.Sprintf(`state == "%s"`, filter.State.String()))
	}

	if len(params) == 0 {
		return fmt.Sprintf("showDeleted=%v", showDeleted)
	}

	return fmt.Sprintf("filter=%s&showDeleted=%v", url.QueryEscape(strings.Join(params, " && ")), showDeleted)
}

// ListProject list all projects.
func (c *client) ListProject(ctx context.Context, filter *api.ProjectFilter) ([]*v1pb.Project, error) {
	res := []*v1pb.Project{}
	pageToken := ""
	startTime := time.Now()
	query := buildProjectQuery(filter)

	for {
		startTimePerPage := time.Now()
		resp, err := c.listProjectPerPage(ctx, query, pageToken, 500)
		if err != nil {
			return nil, err
		}
		res = append(res, resp.Projects...)
		tflog.Debug(ctx, "[list project per page]", map[string]interface{}{
			"count": len(resp.Projects),
			"ms":    time.Since(startTimePerPage).Milliseconds(),
		})

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	tflog.Debug(ctx, "[list project]", map[string]interface{}{
		"total": len(res),
		"ms":    time.Since(startTime).Milliseconds(),
	})

	return res, nil
}

// listProjectPerPage list the projects.
func (c *client) listProjectPerPage(ctx context.Context, query, pageToken string, pageSize int) (*v1pb.ListProjectsResponse, error) {
	requestURL := fmt.Sprintf(
		"%s/%s/projects?%s&page_size=%d&page_token=%s",
		c.url,
		c.version,
		query,
		pageSize,
		url.QueryEscape(pageToken),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
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
