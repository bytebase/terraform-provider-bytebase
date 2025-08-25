package client

import (
	"context"
	"errors"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ListPolicies lists policies in a specific resource.
func (c *client) ListPolicies(ctx context.Context, parent string) (*v1pb.ListPoliciesResponse, error) {
	if c.orgPolicyClient == nil {
		return nil, errors.New("org policy service client not initialized")
	}

	req := connect.NewRequest(&v1pb.ListPoliciesRequest{
		Parent: parent,
	})

	resp, err := c.orgPolicyClient.ListPolicies(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// GetPolicy gets a policy in a specific resource.
func (c *client) GetPolicy(ctx context.Context, policyName string) (*v1pb.Policy, error) {
	if c.orgPolicyClient == nil {
		return nil, errors.New("org policy service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetPolicyRequest{
		Name: policyName,
	})

	resp, err := c.orgPolicyClient.GetPolicy(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UpsertPolicy creates or updates the policy.
func (c *client) UpsertPolicy(ctx context.Context, policy *v1pb.Policy, updateMasks []string) (*v1pb.Policy, error) {
	if c.orgPolicyClient == nil {
		return nil, errors.New("org policy service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdatePolicyRequest{
		Policy: policy,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: updateMasks,
		},
		AllowMissing: true,
	})

	resp, err := c.orgPolicyClient.UpdatePolicy(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// DeletePolicy deletes the policy.
func (c *client) DeletePolicy(ctx context.Context, policyName string) error {
	if c.orgPolicyClient == nil {
		return errors.New("org policy service client not initialized")
	}

	req := connect.NewRequest(&v1pb.DeletePolicyRequest{
		Name: policyName,
	})

	_, err := c.orgPolicyClient.DeletePolicy(ctx, req)
	return err
}
