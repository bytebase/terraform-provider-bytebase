package client

import (
	"context"
	"fmt"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ListDatabaseGroup list all database groups in a project using Connect RPC.
func (c *client) ListDatabaseGroup(ctx context.Context, project string) (*v1pb.ListDatabaseGroupsResponse, error) {
	if c.databaseGroupClient == nil {
		return nil, fmt.Errorf("database group service client not initialized")
	}

	req := connect.NewRequest(&v1pb.ListDatabaseGroupsRequest{
		Parent: project,
	})

	resp, err := c.databaseGroupClient.ListDatabaseGroups(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// CreateDatabaseGroup creates the database group using Connect RPC.
func (c *client) CreateDatabaseGroup(ctx context.Context, project, groupID string, group *v1pb.DatabaseGroup) (*v1pb.DatabaseGroup, error) {
	if c.databaseGroupClient == nil {
		return nil, fmt.Errorf("database group service client not initialized")
	}

	req := connect.NewRequest(&v1pb.CreateDatabaseGroupRequest{
		Parent:           project,
		DatabaseGroupId:  groupID,
		DatabaseGroup:    group,
	})

	resp, err := c.databaseGroupClient.CreateDatabaseGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// GetDatabaseGroup gets the database group by name using Connect RPC.
func (c *client) GetDatabaseGroup(ctx context.Context, name string, view v1pb.DatabaseGroupView) (*v1pb.DatabaseGroup, error) {
	if c.databaseGroupClient == nil {
		return nil, fmt.Errorf("database group service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetDatabaseGroupRequest{
		Name: name,
		View: view,
	})

	resp, err := c.databaseGroupClient.GetDatabaseGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UpdateDatabaseGroup updates the database group using Connect RPC.
func (c *client) UpdateDatabaseGroup(ctx context.Context, patch *v1pb.DatabaseGroup, updateMasks []string) (*v1pb.DatabaseGroup, error) {
	if c.databaseGroupClient == nil {
		return nil, fmt.Errorf("database group service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdateDatabaseGroupRequest{
		DatabaseGroup: patch,
		UpdateMask:    &fieldmaskpb.FieldMask{Paths: updateMasks},
	})

	resp, err := c.databaseGroupClient.UpdateDatabaseGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// DeleteDatabaseGroup deletes the database group.
func (c *client) DeleteDatabaseGroup(ctx context.Context, databaseGroupName string) error {
	if c.databaseGroupClient == nil {
		return fmt.Errorf("database group service client not initialized")
	}

	req := connect.NewRequest(&v1pb.DeleteDatabaseGroupRequest{
		Name: databaseGroupName,
	})

	_, err := c.databaseGroupClient.DeleteDatabaseGroup(ctx, req)
	return err
}
