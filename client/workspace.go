package client

import (
	"context"
	"errors"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// GetWorkspace gets the workspace by name.
func (c *client) GetWorkspace(ctx context.Context, workspaceName string) (*v1pb.Workspace, error) {
	if c.workspaceClient == nil {
		return nil, errors.New("workspace service client not initialized")
	}

	resp, err := c.workspaceClient.GetWorkspace(ctx, connect.NewRequest(&v1pb.GetWorkspaceRequest{
		Name: workspaceName,
	}))
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UpdateWorkspace updates the workspace.
func (c *client) UpdateWorkspace(ctx context.Context, patch *v1pb.Workspace, updateMasks []string) (*v1pb.Workspace, error) {
	if c.workspaceClient == nil {
		return nil, errors.New("workspace service client not initialized")
	}

	resp, err := c.workspaceClient.UpdateWorkspace(ctx, connect.NewRequest(&v1pb.UpdateWorkspaceRequest{
		Workspace:  patch,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: updateMasks},
	}))
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// GetWorkspaceIAMPolicy gets the workspace IAM policy.
func (c *client) GetWorkspaceIAMPolicy(ctx context.Context) (*v1pb.IamPolicy, error) {
	if c.workspaceClient == nil {
		return nil, errors.New("workspace service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetIamPolicyRequest{
		Resource: c.workspaceName,
	})

	resp, err := c.workspaceClient.GetIamPolicy(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// SetWorkspaceIAMPolicy sets the workspace IAM policy.
func (c *client) SetWorkspaceIAMPolicy(ctx context.Context, setIamPolicyRequest *v1pb.SetIamPolicyRequest) (*v1pb.IamPolicy, error) {
	if c.workspaceClient == nil {
		return nil, errors.New("workspace service client not initialized")
	}

	// Ensure the resource is set correctly
	if setIamPolicyRequest.Resource == "" {
		setIamPolicyRequest.Resource = c.workspaceName
	}

	req := connect.NewRequest(setIamPolicyRequest)

	resp, err := c.workspaceClient.SetIamPolicy(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}
