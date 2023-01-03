// Package client is the API message for Bytebase API client.
package client

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// client is the API message for Bytebase API client.
type client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
	Auth       *api.Login
}

// NewClient returns the new Bytebase API client.
func NewClient(url, email, password string) (api.Client, error) {
	c := client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		HostURL:    url,
	}

	c.Auth = &api.Login{
		Email:    email,
		Password: password,
	}

	ar, err := c.Login()
	if err != nil {
		return nil, err
	}

	c.Token = ar.Token

	return &c, nil
}

func (c *client) doRequest(req *http.Request) ([]byte, error) {
	if c.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	}

	res, err := c.HTTPClient.Do(req)
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
