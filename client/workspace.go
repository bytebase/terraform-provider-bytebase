package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// GetWorkspaceIAMPolicy gets the workspace IAM policy.
func (c *client) GetWorkspaceIAMPolicy(ctx context.Context) (*v1pb.IamPolicy, error) {
	body, err := c.getResource(ctx, "workspaces/-:getIamPolicy", "")
	if err != nil {
		return nil, err
	}

	var res v1pb.IamPolicy
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetWorkspaceIAMPolicy sets the workspace IAM policy.
func (c *client) SetWorkspaceIAMPolicy(ctx context.Context, setIamPolicyRequest *v1pb.SetIamPolicyRequest) (*v1pb.IamPolicy, error) {
	payload, err := protojson.Marshal(setIamPolicyRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/%s:setIamPolicy", c.url, c.version, "workspaces/-"), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.IamPolicy
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
