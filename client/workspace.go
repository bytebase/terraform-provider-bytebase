package client

import (
	"context"
	"fmt"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
)

// GetWorkspaceIAMPolicy gets the workspace IAM policy.
func (c *client) GetWorkspaceIAMPolicy(ctx context.Context) (*v1pb.IamPolicy, error) {
	if c.workspaceClient == nil {
		return nil, fmt.Errorf("workspace service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetIamPolicyRequest{
		Resource: "workspaces/-",
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
		return nil, fmt.Errorf("workspace service client not initialized")
	}

	// Ensure the resource is set correctly
	if setIamPolicyRequest.Resource == "" {
		setIamPolicyRequest.Resource = "workspaces/-"
	}

	req := connect.NewRequest(setIamPolicyRequest)

	resp, err := c.workspaceClient.SetIamPolicy(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}
