package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// Login will login the user and get the response.
func (c *Client) Login() (*api.AuthResponse, error) {
	if c.Auth.Email == "" || c.Auth.Password == "" {
		return nil, errors.Errorf("define username and password")
	}
	rb, err := json.Marshal(c.Auth)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/auth/login", c.HostURL), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	ar := api.AuthResponse{}
	err = json.Unmarshal(body, &ar)
	if err != nil {
		return nil, err
	}

	return &ar, nil
}
