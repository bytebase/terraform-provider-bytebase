package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// GetDatabase gets the database by environment resource id, instance resource id and the database name.
func (c *client) GetDatabase(ctx context.Context, find *api.DatabaseFindMessage) (*api.DatabaseMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/instances/%s/databases/%s", c.url, c.version, find.InstanceID, find.DatabaseName), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.DatabaseMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// ListDatabase list the databases.
func (c *client) ListDatabase(ctx context.Context, find *api.DatabaseFindMessage) (*api.ListDatabaseMessage, error) {
	requestURL := fmt.Sprintf("%s/%s/instances/%s/databases", c.url, c.version, find.InstanceID)
	if v := find.Filter; v != nil {
		requestURL = fmt.Sprintf("%s?filter=%s", requestURL, url.QueryEscape(*v))
	}

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.ListDatabaseMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateDatabase patches the database.
func (c *client) UpdateDatabase(ctx context.Context, instanceID, databaseName string, patch *api.DatabasePatchMessage) (*api.DatabaseMessage, error) {
	patch.Name = fmt.Sprintf("instances/%s/databases/%s", instanceID, databaseName)
	payload, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	updateMask := []string{}
	if patch.Project != nil {
		updateMask = append(updateMask, "project")
	}
	if patch.Labels != nil {
		updateMask = append(updateMask, "labels")
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/%s/instances/%s/databases/%s?update_mask=%s", c.url, c.version, instanceID, databaseName, strings.Join(updateMask, ",")), strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.DatabaseMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
