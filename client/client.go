// Package client is the API message for Bytebase API client.
package client

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// Client is the API message for Bytebase API client.
type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
	Auth       *authStruct
}

type authStruct struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// NewClient returns the new Bytebase API client.
func NewClient(url, email, password string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		HostURL:    url,
	}

	c.Auth = &authStruct{
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

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

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
