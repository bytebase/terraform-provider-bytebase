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

// ListServiceAccount list all service accounts using Connect RPC.
func (c *client) ListServiceAccount(ctx context.Context, parent string, showDeleted bool) ([]*v1pb.ServiceAccount, error) {
	if c.serviceAccountClient == nil {
		return nil, errors.New("service account service client not initialized")
	}

	res := []*v1pb.ServiceAccount{}
	pageToken := ""
	startTime := time.Now()

	for {
		startTimePerPage := time.Now()

		req := connect.NewRequest(&v1pb.ListServiceAccountsRequest{
			Parent:      parent,
			PageSize:    500,
			PageToken:   pageToken,
			ShowDeleted: showDeleted,
		})

		resp, err := c.serviceAccountClient.ListServiceAccounts(ctx, req)
		if err != nil {
			return nil, err
		}

		res = append(res, resp.Msg.ServiceAccounts...)

		tflog.Debug(ctx, "[list service account per page]", map[string]interface{}{
			"count": len(resp.Msg.ServiceAccounts),
			"ms":    time.Since(startTimePerPage).Milliseconds(),
		})

		pageToken = resp.Msg.NextPageToken
		if pageToken == "" {
			break
		}
	}

	tflog.Debug(ctx, "[list service account]", map[string]interface{}{
		"total": len(res),
		"ms":    time.Since(startTime).Milliseconds(),
	})

	return res, nil
}

// CreateServiceAccount creates the service account using Connect RPC.
func (c *client) CreateServiceAccount(ctx context.Context, parent, serviceAccountID string, serviceAccount *v1pb.ServiceAccount) (*v1pb.ServiceAccount, error) {
	if c.serviceAccountClient == nil {
		return nil, errors.New("service account service client not initialized")
	}

	req := connect.NewRequest(&v1pb.CreateServiceAccountRequest{
		Parent:           parent,
		ServiceAccountId: serviceAccountID,
		ServiceAccount:   serviceAccount,
	})

	resp, err := c.serviceAccountClient.CreateServiceAccount(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// GetServiceAccount gets the service account by name using Connect RPC.
func (c *client) GetServiceAccount(ctx context.Context, name string) (*v1pb.ServiceAccount, error) {
	if c.serviceAccountClient == nil {
		return nil, errors.New("service account service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetServiceAccountRequest{
		Name: name,
	})

	resp, err := c.serviceAccountClient.GetServiceAccount(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UpdateServiceAccount updates the service account using Connect RPC.
func (c *client) UpdateServiceAccount(ctx context.Context, patch *v1pb.ServiceAccount, updateMasks []string) (*v1pb.ServiceAccount, error) {
	if c.serviceAccountClient == nil {
		return nil, errors.New("service account service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdateServiceAccountRequest{
		ServiceAccount: patch,
		UpdateMask:     &fieldmaskpb.FieldMask{Paths: updateMasks},
	})

	resp, err := c.serviceAccountClient.UpdateServiceAccount(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UndeleteServiceAccount undeletes the service account by name using Connect RPC.
func (c *client) UndeleteServiceAccount(ctx context.Context, name string) (*v1pb.ServiceAccount, error) {
	if c.serviceAccountClient == nil {
		return nil, errors.New("service account service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UndeleteServiceAccountRequest{
		Name: name,
	})

	resp, err := c.serviceAccountClient.UndeleteServiceAccount(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// DeleteServiceAccount deletes the service account.
func (c *client) DeleteServiceAccount(ctx context.Context, name string) error {
	if c.serviceAccountClient == nil {
		return errors.New("service account service client not initialized")
	}

	req := connect.NewRequest(&v1pb.DeleteServiceAccountRequest{
		Name: name,
	})

	_, err := c.serviceAccountClient.DeleteServiceAccount(ctx, req)
	return err
}
