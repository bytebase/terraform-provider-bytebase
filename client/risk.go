package client

import (
	"context"
	"errors"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ListRisk lists the risk using Connect RPC.
func (c *client) ListRisk(ctx context.Context) ([]*v1pb.Risk, error) {
	if c.riskClient == nil {
		return nil, errors.New("risk service client not initialized")
	}

	req := connect.NewRequest(&v1pb.ListRisksRequest{})

	resp, err := c.riskClient.ListRisks(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg.Risks, nil
}

// GetRisk gets the risk by full name using Connect RPC.
func (c *client) GetRisk(ctx context.Context, name string) (*v1pb.Risk, error) {
	if c.riskClient == nil {
		return nil, errors.New("risk service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetRiskRequest{
		Name: name,
	})

	resp, err := c.riskClient.GetRisk(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// CreateRisk creates the risk using Connect RPC.
func (c *client) CreateRisk(ctx context.Context, risk *v1pb.Risk) (*v1pb.Risk, error) {
	if c.riskClient == nil {
		return nil, errors.New("risk service client not initialized")
	}

	req := connect.NewRequest(&v1pb.CreateRiskRequest{
		Risk: risk,
	})

	resp, err := c.riskClient.CreateRisk(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UpdateRisk updates the risk using Connect RPC.
func (c *client) UpdateRisk(ctx context.Context, patch *v1pb.Risk, updateMasks []string) (*v1pb.Risk, error) {
	if c.riskClient == nil {
		return nil, errors.New("risk service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdateRiskRequest{
		Risk:       patch,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: updateMasks},
	})

	resp, err := c.riskClient.UpdateRisk(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// DeleteRisk deletes the risk.
func (c *client) DeleteRisk(ctx context.Context, name string) error {
	if c.riskClient == nil {
		return errors.New("risk service client not initialized")
	}

	req := connect.NewRequest(&v1pb.DeleteRiskRequest{
		Name: name,
	})

	_, err := c.riskClient.DeleteRisk(ctx, req)
	return err
}
