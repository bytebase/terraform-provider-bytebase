package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func buildInstanceFilter(filter *api.InstanceFilter) string {
	params := []string{}

	if v := filter.Query; v != "" {
		params = append(params, fmt.Sprintf(`(name.matches("%s") || resource_id.matches("%s"))`, strings.ToLower(v), strings.ToLower(v)))
	}
	if v := filter.Project; v != "" {
		params = append(params, fmt.Sprintf(`project == "%s"`, v))
	}
	if v := filter.Environment; v != "" {
		params = append(params, fmt.Sprintf(`environment == "%s"`, v))
	}
	if v := filter.Host; v != "" {
		params = append(params, fmt.Sprintf(`host == "%s"`, v))
	}
	if v := filter.Port; v != "" {
		params = append(params, fmt.Sprintf(`port == "%s"`, v))
	}
	if v := filter.Engines; len(v) > 0 {
		engines := []string{}
		for _, e := range v {
			engines = append(engines, fmt.Sprintf(`"%s"`, e.String()))
		}
		params = append(params, fmt.Sprintf(`engine in [%s]`, strings.Join(engines, ", ")))
	}
	if filter.State == v1pb.State_DELETED {
		params = append(params, fmt.Sprintf(`state == "%s"`, filter.State.String()))
	}

	return strings.Join(params, " && ")
}

// ListInstance will return instances using Connect RPC.
func (c *client) ListInstance(ctx context.Context, filter *api.InstanceFilter) ([]*v1pb.Instance, error) {
	if c.instanceClient == nil {
		return nil, errors.New("instance service client not initialized")
	}

	res := []*v1pb.Instance{}
	pageToken := ""
	startTime := time.Now()
	filterStr := buildInstanceFilter(filter)
	showDeleted := filter.State == v1pb.State_DELETED

	for {
		startTimePerPage := time.Now()

		req := connect.NewRequest(&v1pb.ListInstancesRequest{
			Filter:      filterStr,
			PageSize:    500,
			PageToken:   pageToken,
			ShowDeleted: showDeleted,
		})

		resp, err := c.instanceClient.ListInstances(ctx, req)
		if err != nil {
			return nil, err
		}

		res = append(res, resp.Msg.Instances...)

		tflog.Debug(ctx, "[list instance per page]", map[string]interface{}{
			"count": len(resp.Msg.Instances),
			"ms":    time.Since(startTimePerPage).Milliseconds(),
		})

		pageToken = resp.Msg.NextPageToken
		if pageToken == "" {
			break
		}
	}

	tflog.Debug(ctx, "[list instance]", map[string]interface{}{
		"total": len(res),
		"ms":    time.Since(startTime).Milliseconds(),
	})

	return res, nil
}

// GetInstance gets the instance by full name using Connect RPC.
func (c *client) GetInstance(ctx context.Context, instanceName string) (*v1pb.Instance, error) {
	if c.instanceClient == nil {
		return nil, errors.New("instance service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetInstanceRequest{
		Name: instanceName,
	})

	resp, err := c.instanceClient.GetInstance(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// CreateInstance creates the instance using Connect RPC.
func (c *client) CreateInstance(ctx context.Context, instanceID string, instance *v1pb.Instance) (*v1pb.Instance, error) {
	if c.instanceClient == nil {
		return nil, errors.New("instance service client not initialized")
	}

	req := connect.NewRequest(&v1pb.CreateInstanceRequest{
		InstanceId: instanceID,
		Instance:   instance,
	})

	resp, err := c.instanceClient.CreateInstance(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UpdateInstance updates the instance using Connect RPC.
func (c *client) UpdateInstance(ctx context.Context, patch *v1pb.Instance, updateMasks []string) (*v1pb.Instance, error) {
	if c.instanceClient == nil {
		return nil, errors.New("instance service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdateInstanceRequest{
		Instance:   patch,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: updateMasks},
	})

	resp, err := c.instanceClient.UpdateInstance(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UndeleteInstance undeletes the instance using Connect RPC.
func (c *client) UndeleteInstance(ctx context.Context, instanceName string) (*v1pb.Instance, error) {
	if c.instanceClient == nil {
		return nil, errors.New("instance service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UndeleteInstanceRequest{
		Name: instanceName,
	})

	resp, err := c.instanceClient.UndeleteInstance(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// SyncInstanceSchema will trigger the schema sync for an instance using Connect RPC.
func (c *client) SyncInstanceSchema(ctx context.Context, instanceName string) error {
	if c.instanceClient == nil {
		return errors.New("instance service client not initialized")
	}

	req := connect.NewRequest(&v1pb.SyncInstanceRequest{
		Name: instanceName,
	})

	_, err := c.instanceClient.SyncInstance(ctx, req)
	return err
}

// DeleteInstance deletes the instance.
func (c *client) DeleteInstance(ctx context.Context, name string) error {
	if c.instanceClient == nil {
		return errors.New("instance service client not initialized")
	}

	req := connect.NewRequest(&v1pb.DeleteInstanceRequest{
		Name: name,
	})

	_, err := c.instanceClient.DeleteInstance(ctx, req)
	return err
}
