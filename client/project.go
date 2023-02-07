package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// GetProject gets the project by resource id.
func (c *client) GetProject(ctx context.Context, projectID string) (*api.PorjectMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/projects/%s", c.url, c.version, projectID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res api.PorjectMessage
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
