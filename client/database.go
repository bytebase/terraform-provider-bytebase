package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
)

// GetDatabase gets the database by the database full name.
func (c *client) GetDatabase(ctx context.Context, databaseName string) (*v1pb.Database, error) {
	body, err := c.getResource(ctx, databaseName)
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
	body, err := c.updateResource(ctx, patch.Name, patch, updateMasks, false /* allow missing = false*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.Database
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetDatabaseCatalog gets the database catalog by the database full name.
func (c *client) GetDatabaseCatalog(ctx context.Context, databaseName string) (*v1pb.DatabaseCatalog, error) {
	body, err := c.getResource(ctx, fmt.Sprintf("%s/catalog", databaseName))
	if err != nil {
		return nil, err
	}

	var res v1pb.DatabaseCatalog
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateDatabaseCatalog patches the database catalog.
func (c *client) UpdateDatabaseCatalog(ctx context.Context, patch *v1pb.DatabaseCatalog, updateMasks []string) (*v1pb.DatabaseCatalog, error) {
	body, err := c.updateResource(ctx, patch.Name, patch, updateMasks, false /* allow missing = false*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.DatabaseCatalog
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
