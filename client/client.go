// Package client is the API message for Bytebase API client.
package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// client is the API message for Bytebase API client.
type client struct {
	url     string
	version string
	client  *http.Client
	token   string
	caller  *v1pb.User
}

// NewClient returns the new Bytebase API client.
func NewClient(url, version, email, password string) (api.Client, error) {
	c := client{
		client:  &http.Client{Timeout: 10 * time.Second},
		url:     url,
		version: version,
	}

	response, err := c.login(&v1pb.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	c.token = response.Token
	c.caller = response.User

	return &c, nil
}

func (c *client) doRequest(req *http.Request) ([]byte, error) {
	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}

// GetCaller returns the API caller.
func (c *client) GetCaller() *v1pb.User {
	return c.caller
}

// CheckResourceExist check if the resource exists.
func (c *client) CheckResourceExist(ctx context.Context, name string) error {
	if _, err := c.getResource(ctx, name, ""); err != nil {
		return err
	}
	return nil
}

// DeleteResource force delete the resource by name.
func (c *client) DeleteResource(ctx context.Context, name string) error {
	return c.execDelete(ctx, name)
}
