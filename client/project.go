package client

import (
	"context"
	"errors"
	"strings"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// GetProject gets the project by project full name.
func (c *client) GetProject(ctx context.Context, projectName string) (*v1pb.Project, error) {
	if c.projectClient == nil {
		return nil, errors.New("project service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetProjectRequest{
		Name: projectName,
	})

	resp, err := c.projectClient.GetProject(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// GetProjectIAMPolicy gets the project IAM policy by project full name.
func (c *client) GetProjectIAMPolicy(ctx context.Context, projectName string) (*v1pb.IamPolicy, error) {
	if c.projectClient == nil {
		return nil, errors.New("project service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetIamPolicyRequest{
		Resource: projectName,
	})

	resp, err := c.projectClient.GetIamPolicy(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// SetProjectIAMPolicy sets the project IAM policy.
func (c *client) SetProjectIAMPolicy(ctx context.Context, projectName string, update *v1pb.SetIamPolicyRequest) (*v1pb.IamPolicy, error) {
	if c.projectClient == nil {
		return nil, errors.New("project service client not initialized")
	}

	// Update the resource field to match the project name
	update.Resource = projectName

	req := connect.NewRequest(update)

	resp, err := c.projectClient.SetIamPolicy(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// CreateProjectWebhook creates the webhook in the project.
func (c *client) CreateProjectWebhook(ctx context.Context, projectName string, webhook *v1pb.Webhook) (*v1pb.Webhook, error) {
	if c.projectClient == nil {
		return nil, errors.New("project service client not initialized")
	}

	req := connect.NewRequest(&v1pb.AddWebhookRequest{
		Project: projectName,
		Webhook: webhook,
	})

	resp, err := c.projectClient.AddWebhook(ctx, req)
	if err != nil {
		return nil, err
	}

	// AddWebhook returns the updated project, find the webhook we just added
	if resp.Msg != nil && len(resp.Msg.Webhooks) > 0 {
		// Return the last webhook (the one we just added)
		return resp.Msg.Webhooks[len(resp.Msg.Webhooks)-1], nil
	}

	return nil, errors.New("webhook not found in response")
}

// UpdateProjectWebhook updates the webhook.
func (c *client) UpdateProjectWebhook(ctx context.Context, patch *v1pb.Webhook, updateMasks []string) (*v1pb.Webhook, error) {
	if c.projectClient == nil {
		return nil, errors.New("project service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdateWebhookRequest{
		Webhook: patch,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: updateMasks,
		},
	})

	resp, err := c.projectClient.UpdateWebhook(ctx, req)
	if err != nil {
		return nil, err
	}

	// UpdateWebhook returns the updated project, find the updated webhook
	if resp.Msg != nil && resp.Msg.Webhooks != nil {
		for _, wh := range resp.Msg.Webhooks {
			if wh.Name == patch.Name {
				return wh, nil
			}
		}
	}

	return nil, errors.New("updated webhook not found in response")
}

// DeleteProjectWebhook deletes the webhook.
func (c *client) DeleteProjectWebhook(ctx context.Context, webhookName string) error {
	if c.projectClient == nil {
		return errors.New("project service client not initialized")
	}

	req := connect.NewRequest(&v1pb.RemoveWebhookRequest{
		Webhook: &v1pb.Webhook{
			Name: webhookName,
		},
	})

	_, err := c.projectClient.RemoveWebhook(ctx, req)
	return err
}

// ListProject list all projects.
func (c *client) ListProject(ctx context.Context, filter *api.ProjectFilter) ([]*v1pb.Project, error) {
	if c.projectClient == nil {
		return nil, errors.New("project service client not initialized")
	}

	var projects []*v1pb.Project
	pageToken := ""

	for {
		req := connect.NewRequest(&v1pb.ListProjectsRequest{
			PageSize:    500,
			PageToken:   pageToken,
			ShowDeleted: filter.State == v1pb.State_DELETED,
		})

		resp, err := c.projectClient.ListProjects(ctx, req)
		if err != nil {
			return nil, err
		}

		// Filter projects based on the filter criteria
		for _, project := range resp.Msg.Projects {
			// Apply filter logic
			if filter.Query != "" {
				// Check if name or resource_id matches the query
				// Extract resource ID from name (e.g., "projects/my-project" -> "my-project")
				resourceID := ""
				if parts := strings.Split(project.Name, "/"); len(parts) > 1 {
					resourceID = parts[len(parts)-1]
				}
				if !containsIgnoreCase(project.Name, filter.Query) && !containsIgnoreCase(resourceID, filter.Query) {
					continue
				}
			}

			if filter.ExcludeDefault {
				// Skip default project (you might need to adjust this logic)
				if project.Name == "projects/default" {
					continue
				}
			}

			if filter.State != v1pb.State_STATE_UNSPECIFIED && project.State != filter.State {
				continue
			}

			projects = append(projects, project)
		}

		pageToken = resp.Msg.NextPageToken
		if pageToken == "" {
			break
		}
	}

	tflog.Debug(ctx, "[list project]", map[string]interface{}{
		"total": len(projects),
	})

	return projects, nil
}

// CreateProject creates the project.
func (c *client) CreateProject(ctx context.Context, projectID string, project *v1pb.Project) (*v1pb.Project, error) {
	if c.projectClient == nil {
		return nil, errors.New("project service client not initialized")
	}

	req := connect.NewRequest(&v1pb.CreateProjectRequest{
		ProjectId: projectID,
		Project:   project,
	})

	resp, err := c.projectClient.CreateProject(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UpdateProject updates the project.
func (c *client) UpdateProject(ctx context.Context, patch *v1pb.Project, updateMasks []string) (*v1pb.Project, error) {
	if c.projectClient == nil {
		return nil, errors.New("project service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdateProjectRequest{
		Project: patch,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: updateMasks,
		},
	})

	resp, err := c.projectClient.UpdateProject(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UndeleteProject undeletes the project.
func (c *client) UndeleteProject(ctx context.Context, projectName string) (*v1pb.Project, error) {
	if c.projectClient == nil {
		return nil, errors.New("project service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UndeleteProjectRequest{
		Name: projectName,
	})

	resp, err := c.projectClient.UndeleteProject(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// DeleteProject deletes the project.
func (c *client) DeleteProject(ctx context.Context, name string) error {
	if c.projectClient == nil {
		return errors.New("project service client not initialized")
	}

	req := connect.NewRequest(&v1pb.DeleteProjectRequest{
		Name: name,
	})

	_, err := c.projectClient.DeleteProject(ctx, req)
	return err
}

// Helper function for case-insensitive string contains.
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
