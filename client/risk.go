package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// ListRisk lists the risk.
func (c *client) ListRisk(ctx context.Context) ([]*v1pb.Risk, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/risks", c.url, c.version), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.ListRisksResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return res.Risks, nil
}

// GetRisk gets the risk by full name.
func (c *client) GetRisk(ctx context.Context, name string) (*v1pb.Risk, error) {
	body, err := c.getResource(ctx, name, "")
	if err != nil {
		return nil, err
	}

	var res v1pb.Risk
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// CreateRisk creates the risk.
func (c *client) CreateRisk(ctx context.Context, risk *v1pb.Risk) (*v1pb.Risk, error) {
	payload, err := protojson.Marshal(risk)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/risks", c.url, c.version), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.Risk
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateRisk updates the risk.
func (c *client) UpdateRisk(ctx context.Context, patch *v1pb.Risk, updateMasks []string) (*v1pb.Risk, error) {
	body, err := c.updateResource(ctx, patch.Name, patch, updateMasks, false /* allow missing = false*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.Risk
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
