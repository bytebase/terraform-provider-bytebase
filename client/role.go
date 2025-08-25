package client

import (
	"context"
	"fmt"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// GetRole gets the role by full name using Connect RPC.
func (c *client) GetRole(ctx context.Context, name string) (*v1pb.Role, error) {
	if c.roleClient == nil {
		return nil, fmt.Errorf("role service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetRoleRequest{
		Name: name,
	})

	resp, err := c.roleClient.GetRole(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// CreateRole creates the role using Connect RPC.
func (c *client) CreateRole(ctx context.Context, roleID string, role *v1pb.Role) (*v1pb.Role, error) {
	if c.roleClient == nil {
		return nil, fmt.Errorf("role service client not initialized")
	}

	req := connect.NewRequest(&v1pb.CreateRoleRequest{
		Role:   role,
		RoleId: roleID,
	})

	resp, err := c.roleClient.CreateRole(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UpdateRole updates the role using Connect RPC.
func (c *client) UpdateRole(ctx context.Context, patch *v1pb.Role, updateMasks []string) (*v1pb.Role, error) {
	if c.roleClient == nil {
		return nil, fmt.Errorf("role service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdateRoleRequest{
		Role:         patch,
		AllowMissing: true,
		UpdateMask:   &fieldmaskpb.FieldMask{Paths: updateMasks},
	})

	resp, err := c.roleClient.UpdateRole(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// ListRole will returns all roles using Connect RPC.
func (c *client) ListRole(ctx context.Context) (*v1pb.ListRolesResponse, error) {
	if c.roleClient == nil {
		return nil, fmt.Errorf("role service client not initialized")
	}

	req := connect.NewRequest(&v1pb.ListRolesRequest{})

	resp, err := c.roleClient.ListRoles(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// DeleteRole deletes the role.
func (c *client) DeleteRole(ctx context.Context, name string) error {
	if c.roleClient == nil {
		return fmt.Errorf("role service client not initialized")
	}

	req := connect.NewRequest(&v1pb.DeleteRoleRequest{
		Name: name,
	})

	_, err := c.roleClient.DeleteRole(ctx, req)
	return err
}
