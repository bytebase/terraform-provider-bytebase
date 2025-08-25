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

func buildGroupFilter(filter *api.GroupFilter) string {
	params := []string{}

	if v := filter.Query; v != "" {
		params = append(params, fmt.Sprintf(`(title.matches("%s") || email.matches("%s"))`, strings.ToLower(v), strings.ToLower(v)))
	}
	if v := filter.Project; v != "" {
		params = append(params, fmt.Sprintf(`project == "%s"`, v))
	}

	return strings.Join(params, " && ")
}

// ListGroup list all groups using Connect RPC.
func (c *client) ListGroup(ctx context.Context, filter *api.GroupFilter) ([]*v1pb.Group, error) {
	if c.groupClient == nil {
		return nil, fmt.Errorf("group service client not initialized")
	}

	res := []*v1pb.Group{}
	pageToken := ""
	startTime := time.Now()
	filterStr := buildGroupFilter(filter)

	for {
		startTimePerPage := time.Now()

		req := connect.NewRequest(&v1pb.ListGroupsRequest{
			PageSize:  500,
			PageToken: pageToken,
			Filter:    filterStr,
		})
		resp, err := c.groupClient.ListGroups(ctx, req)
		if err != nil {
			return nil, err
		}

		res = append(res, resp.Msg.Groups...)

		tflog.Debug(ctx, "[list group per page]", map[string]interface{}{
			"count": len(resp.Msg.Groups),
			"ms":    time.Since(startTimePerPage).Milliseconds(),
		})

		pageToken = resp.Msg.NextPageToken
		if pageToken == "" {
			break
		}
	}

	tflog.Debug(ctx, "[list group]", map[string]interface{}{
		"total": len(res),
		"ms":    time.Since(startTime).Milliseconds(),
	})

	return res, nil
}

// CreateGroup creates the group using Connect RPC.
func (c *client) CreateGroup(ctx context.Context, email string, group *v1pb.Group) (*v1pb.Group, error) {
	if c.groupClient == nil {
		return nil, fmt.Errorf("group service client not initialized")
	}

	req := connect.NewRequest(&v1pb.CreateGroupRequest{
		Group:      group,
		GroupEmail: email,
	})

	resp, err := c.groupClient.CreateGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// GetGroup gets the group by name using Connect RPC.
func (c *client) GetGroup(ctx context.Context, name string) (*v1pb.Group, error) {
	if c.groupClient == nil {
		return nil, fmt.Errorf("group service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetGroupRequest{
		Name: name,
	})

	resp, err := c.groupClient.GetGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UpdateGroup updates the group using Connect RPC.
func (c *client) UpdateGroup(ctx context.Context, patch *v1pb.Group, updateMasks []string) (*v1pb.Group, error) {
	if c.groupClient == nil {
		return nil, fmt.Errorf("group service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdateGroupRequest{
		Group:        patch,
		AllowMissing: true,
		UpdateMask:   &fieldmaskpb.FieldMask{Paths: updateMasks},
	})

	resp, err := c.groupClient.UpdateGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// DeleteGroup deletes the group.
func (c *client) DeleteGroup(ctx context.Context, name string) error {
	if c.groupClient == nil {
		return fmt.Errorf("group service client not initialized")
	}

	req := connect.NewRequest(&v1pb.DeleteGroupRequest{
		Name: name,
	})

	_, err := c.groupClient.DeleteGroup(ctx, req)
	return err
}
