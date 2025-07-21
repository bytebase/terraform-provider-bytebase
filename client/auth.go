package client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
)

// Login will login the user and get the response.
func (c *client) login(request *v1pb.LoginRequest) (*v1pb.LoginResponse, error) {
	if request.Email == "" || request.Password == "" {
		return nil, errors.Errorf("undefined login service account or key")
	}
	rb, err := protojson.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/auth/login", c.url, c.version), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to login")
	}

	ar := v1pb.LoginResponse{}
	if err := ProtojsonUnmarshaler.Unmarshal(body, &ar); err != nil {
		return nil, err
	}

	return &ar, nil
}
