package client

import (
	"context"
	"fmt"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ListSettings lists all settings using Connect RPC.
func (c *client) ListSettings(ctx context.Context) (*v1pb.ListSettingsResponse, error) {
	if c.settingClient == nil {
		return nil, fmt.Errorf("setting service client not initialized")
	}

	req := connect.NewRequest(&v1pb.ListSettingsRequest{})

	resp, err := c.settingClient.ListSettings(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// GetSetting gets the setting by the name using Connect RPC.
func (c *client) GetSetting(ctx context.Context, settingName string) (*v1pb.Setting, error) {
	if c.settingClient == nil {
		return nil, fmt.Errorf("setting service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetSettingRequest{
		Name: settingName,
	})

	resp, err := c.settingClient.GetSetting(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UpsertSetting updates or creates the setting using Connect RPC.
func (c *client) UpsertSetting(ctx context.Context, upsert *v1pb.Setting, updateMasks []string) (*v1pb.Setting, error) {
	if c.settingClient == nil {
		return nil, fmt.Errorf("setting service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdateSettingRequest{
		Setting:      upsert,
		AllowMissing: true,
		UpdateMask:   &fieldmaskpb.FieldMask{Paths: updateMasks},
	})

	resp, err := c.settingClient.UpdateSetting(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}
