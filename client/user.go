package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func buildUserFilter(filter *api.UserFilter) string {
	params := []string{}

	if v := filter.Name; v != "" {
		params = append(params, fmt.Sprintf(`name == "%s"`, strings.ToLower(v)))
	}
	if v := filter.Email; v != "" {
		params = append(params, fmt.Sprintf(`email == "%s"`, strings.ToLower(v)))
	}
	if v := filter.Project; v != "" {
		params = append(params, fmt.Sprintf(`project == "%s"`, v))
	}
	if v := filter.UserTypes; len(v) > 0 {
		userTypes := []string{}
		for _, t := range v {
			userTypes = append(userTypes, fmt.Sprintf(`"%s"`, t.String()))
		}
		params = append(params, fmt.Sprintf(`user_type in [%s]`, strings.Join(userTypes, ", ")))
	}
	if filter.State == v1pb.State_DELETED {
		params = append(params, fmt.Sprintf(`state == "%s"`, filter.State.String()))
	}

	return strings.Join(params, " && ")
}

// ListUser list all users using Connect RPC.
func (c *client) ListUser(ctx context.Context, filter *api.UserFilter) ([]*v1pb.User, error) {
	if c.userClient == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	res := []*v1pb.User{}
	pageToken := ""
	startTime := time.Now()
	filterStr := buildUserFilter(filter)
	showDeleted := filter.State == v1pb.State_DELETED

	for {
		startTimePerPage := time.Now()
		
		req := connect.NewRequest(&v1pb.ListUsersRequest{
			Filter:      filterStr,
			PageSize:    500,
			PageToken:   pageToken,
			ShowDeleted: showDeleted,
		})

		resp, err := c.userClient.ListUsers(ctx, req)
		if err != nil {
			return nil, err
		}

		res = append(res, resp.Msg.Users...)

		tflog.Debug(ctx, "[list user per page]", map[string]interface{}{
			"count": len(resp.Msg.Users),
			"ms":    time.Since(startTimePerPage).Milliseconds(),
		})

		pageToken = resp.Msg.NextPageToken
		if pageToken == "" {
			break
		}
	}

	tflog.Debug(ctx, "[list user]", map[string]interface{}{
		"total": len(res),
		"ms":    time.Since(startTime).Milliseconds(),
	})

	return res, nil
}

// CreateUser creates the user using Connect RPC.
func (c *client) CreateUser(ctx context.Context, user *v1pb.User) (*v1pb.User, error) {
	if c.userClient == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	req := connect.NewRequest(&v1pb.CreateUserRequest{
		User: user,
	})

	resp, err := c.userClient.CreateUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// GetUser gets the user by name using Connect RPC.
func (c *client) GetUser(ctx context.Context, userName string) (*v1pb.User, error) {
	if c.userClient == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetUserRequest{
		Name: userName,
	})

	resp, err := c.userClient.GetUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UpdateUser updates the user using Connect RPC.
func (c *client) UpdateUser(ctx context.Context, patch *v1pb.User, updateMasks []string) (*v1pb.User, error) {
	if c.userClient == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdateUserRequest{
		User:       patch,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: updateMasks},
	})

	resp, err := c.userClient.UpdateUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UndeleteUser undeletes the user by name using Connect RPC.
func (c *client) UndeleteUser(ctx context.Context, userName string) (*v1pb.User, error) {
	if c.userClient == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UndeleteUserRequest{
		Name: userName,
	})

	resp, err := c.userClient.UndeleteUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// DeleteUser deletes the user.
func (c *client) DeleteUser(ctx context.Context, name string) error {
	if c.userClient == nil {
		return fmt.Errorf("user service client not initialized")
	}

	req := connect.NewRequest(&v1pb.DeleteUserRequest{
		Name: name,
	})

	_, err := c.userClient.DeleteUser(ctx, req)
	return err
}