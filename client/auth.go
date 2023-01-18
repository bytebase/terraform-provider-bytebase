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
func (c *client) Login() (*api.AuthResponse, error) {
	if c.auth.Email == "" || c.auth.Password == "" {
		return nil, errors.Errorf("define username and password")
	}
	rb, err := json.Marshal(c.auth)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/auth/login", c.url, c.version), strings.NewReader(string(rb)))
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
