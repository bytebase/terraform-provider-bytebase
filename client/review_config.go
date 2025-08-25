package client

import (
	"context"
	"fmt"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ListReviewConfig will return review configs using Connect RPC.
func (c *client) ListReviewConfig(ctx context.Context) (*v1pb.ListReviewConfigsResponse, error) {
	if c.reviewConfigClient == nil {
		return nil, fmt.Errorf("review config service client not initialized")
	}

	req := connect.NewRequest(&v1pb.ListReviewConfigsRequest{})

	resp, err := c.reviewConfigClient.ListReviewConfigs(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// GetReviewConfig gets the review config by full name using Connect RPC.
func (c *client) GetReviewConfig(ctx context.Context, reviewName string) (*v1pb.ReviewConfig, error) {
	if c.reviewConfigClient == nil {
		return nil, fmt.Errorf("review config service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetReviewConfigRequest{
		Name: reviewName,
	})

	resp, err := c.reviewConfigClient.GetReviewConfig(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UpsertReviewConfig updates or creates the review config using Connect RPC.
func (c *client) UpsertReviewConfig(ctx context.Context, patch *v1pb.ReviewConfig, updateMasks []string) (*v1pb.ReviewConfig, error) {
	if c.reviewConfigClient == nil {
		return nil, fmt.Errorf("review config service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdateReviewConfigRequest{
		ReviewConfig: patch,
		AllowMissing: true,
		UpdateMask:   &fieldmaskpb.FieldMask{Paths: updateMasks},
	})

	resp, err := c.reviewConfigClient.UpdateReviewConfig(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// DeleteReviewConfig deletes the review config.
func (c *client) DeleteReviewConfig(ctx context.Context, name string) error {
	if c.reviewConfigClient == nil {
		return fmt.Errorf("review config service client not initialized")
	}

	req := connect.NewRequest(&v1pb.DeleteReviewConfigRequest{
		Name: name,
	})

	_, err := c.reviewConfigClient.DeleteReviewConfig(ctx, req)
	return err
}
