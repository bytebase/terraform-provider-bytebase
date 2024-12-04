package client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
)

// Login will login the user and get the response.
func (c *client) Login() (*v1pb.LoginResponse, error) {
	if c.auth.Email == "" || c.auth.Password == "" {
		return nil, errors.Errorf("define username and password")
	}
	rb, err := protojson.Marshal(c.auth)
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

	ar := v1pb.LoginResponse{}
	err = ProtojsonUnmarshaler.Unmarshal(body, &ar)
	if err != nil {
		return nil, err
	}

	return &ar, nil
}
