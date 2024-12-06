package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// GetDatabase gets the database by environment resource id, instance resource id and the database name.
func (c *client) GetDatabase(ctx context.Context, databaseName string) (*v1pb.Database, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/%s", c.url, c.version, databaseName), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.Database
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ListDatabase list the databases.
func (c *client) ListDatabase(ctx context.Context, instanceID, filter string) (*v1pb.ListDatabasesResponse, error) {
	requestURL := fmt.Sprintf("%s/%s/instances/%s/databases", c.url, c.version, instanceID)
	if filter != "" {
		requestURL = fmt.Sprintf("%s?filter=%s", requestURL, url.QueryEscape(filter))
	}

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.ListDatabasesResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateDatabase patches the database.
func (c *client) UpdateDatabase(ctx context.Context, patch *v1pb.Database, updateMasks []string) (*v1pb.Database, error) {
	payload, err := protojson.Marshal(patch)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/%s/%s?update_mask=%s", c.url, c.version, patch.Name, strings.Join(updateMasks, ",")), strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.Database
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
