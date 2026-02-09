package client

import (
	"context"
	"errors"
	"time"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ListWorkloadIdentity list all workload identities using Connect RPC.
func (c *client) ListWorkloadIdentity(ctx context.Context, parent string, showDeleted bool) ([]*v1pb.WorkloadIdentity, error) {
	if c.workloadIdentityClient == nil {
		return nil, errors.New("workload identity service client not initialized")
	}

	res := []*v1pb.WorkloadIdentity{}
	pageToken := ""
	startTime := time.Now()

	for {
		startTimePerPage := time.Now()

		req := connect.NewRequest(&v1pb.ListWorkloadIdentitiesRequest{
			Parent:      parent,
			PageSize:    500,
			PageToken:   pageToken,
			ShowDeleted: showDeleted,
		})

		resp, err := c.workloadIdentityClient.ListWorkloadIdentities(ctx, req)
		if err != nil {
			return nil, err
		}

		res = append(res, resp.Msg.WorkloadIdentities...)

		tflog.Debug(ctx, "[list workload identity per page]", map[string]interface{}{
			"count": len(resp.Msg.WorkloadIdentities),
			"ms":    time.Since(startTimePerPage).Milliseconds(),
		})

		pageToken = resp.Msg.NextPageToken
		if pageToken == "" {
			break
		}
	}

	tflog.Debug(ctx, "[list workload identity]", map[string]interface{}{
		"total": len(res),
		"ms":    time.Since(startTime).Milliseconds(),
	})

	return res, nil
}

// CreateWorkloadIdentity creates the workload identity using Connect RPC.
func (c *client) CreateWorkloadIdentity(ctx context.Context, parent, workloadIdentityID string, workloadIdentity *v1pb.WorkloadIdentity) (*v1pb.WorkloadIdentity, error) {
	if c.workloadIdentityClient == nil {
		return nil, errors.New("workload identity service client not initialized")
	}

	req := connect.NewRequest(&v1pb.CreateWorkloadIdentityRequest{
		Parent:             parent,
		WorkloadIdentityId: workloadIdentityID,
		WorkloadIdentity:   workloadIdentity,
	})

	resp, err := c.workloadIdentityClient.CreateWorkloadIdentity(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// GetWorkloadIdentity gets the workload identity by name using Connect RPC.
func (c *client) GetWorkloadIdentity(ctx context.Context, name string) (*v1pb.WorkloadIdentity, error) {
	if c.workloadIdentityClient == nil {
		return nil, errors.New("workload identity service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetWorkloadIdentityRequest{
		Name: name,
	})

	resp, err := c.workloadIdentityClient.GetWorkloadIdentity(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UpdateWorkloadIdentity updates the workload identity using Connect RPC.
func (c *client) UpdateWorkloadIdentity(ctx context.Context, patch *v1pb.WorkloadIdentity, updateMasks []string) (*v1pb.WorkloadIdentity, error) {
	if c.workloadIdentityClient == nil {
		return nil, errors.New("workload identity service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdateWorkloadIdentityRequest{
		WorkloadIdentity: patch,
		UpdateMask:       &fieldmaskpb.FieldMask{Paths: updateMasks},
	})

	resp, err := c.workloadIdentityClient.UpdateWorkloadIdentity(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UndeleteWorkloadIdentity undeletes the workload identity by name using Connect RPC.
func (c *client) UndeleteWorkloadIdentity(ctx context.Context, name string) (*v1pb.WorkloadIdentity, error) {
	if c.workloadIdentityClient == nil {
		return nil, errors.New("workload identity service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UndeleteWorkloadIdentityRequest{
		Name: name,
	})

	resp, err := c.workloadIdentityClient.UndeleteWorkloadIdentity(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// DeleteWorkloadIdentity deletes the workload identity.
func (c *client) DeleteWorkloadIdentity(ctx context.Context, name string) error {
	if c.workloadIdentityClient == nil {
		return errors.New("workload identity service client not initialized")
	}

	req := connect.NewRequest(&v1pb.DeleteWorkloadIdentityRequest{
		Name: name,
	})

	_, err := c.workloadIdentityClient.DeleteWorkloadIdentity(ctx, req)
	return err
}
