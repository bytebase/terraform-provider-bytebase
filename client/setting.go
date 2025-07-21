package client

import (
	"context"
	"fmt"
	"net/http"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
)

// ListSettings lists all settings.
func (c *client) ListSettings(ctx context.Context) (*v1pb.ListSettingsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/settings", c.url, c.version), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.ListSettingsResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetSetting gets the setting by the name.
func (c *client) GetSetting(ctx context.Context, settingName string) (*v1pb.Setting, error) {
	body, err := c.getResource(ctx, settingName, "")
	if err != nil {
		return nil, err
	}

	var res v1pb.Setting
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpsertSetting updates or creates the setting.
func (c *client) UpsertSetting(ctx context.Context, upsert *v1pb.Setting, updateMasks []string) (*v1pb.Setting, error) {
	body, err := c.updateResource(ctx, upsert.Name, upsert, updateMasks, true /* allow missing = true*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.Setting
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
