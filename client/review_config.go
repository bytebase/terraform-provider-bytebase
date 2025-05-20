package client

import (
	"context"
	"fmt"
	"net/http"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
)

// ListReviewConfig will return review configs.
func (c *client) ListReviewConfig(ctx context.Context) (*v1pb.ListReviewConfigsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/reviewConfigs", c.url, c.version), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.ListReviewConfigsResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetReviewConfig gets the review config by full name.
func (c *client) GetReviewConfig(ctx context.Context, reviewName string) (*v1pb.ReviewConfig, error) {
	body, err := c.getResource(ctx, reviewName, "")
	if err != nil {
		return nil, err
	}

	var res v1pb.ReviewConfig
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpsertReviewConfig updates or creates the review config.
func (c *client) UpsertReviewConfig(ctx context.Context, patch *v1pb.ReviewConfig, updateMasks []string) (*v1pb.ReviewConfig, error) {
	body, err := c.updateResource(ctx, patch.Name, patch, updateMasks, true /* allow missing */)
	if err != nil {
		return nil, err
	}

	var res v1pb.ReviewConfig
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// DeleteReviewConfig deletes the review config.
func (c *client) DeleteReviewConfig(ctx context.Context, reviewName string) error {
	return c.deleteResource(ctx, reviewName)
}
