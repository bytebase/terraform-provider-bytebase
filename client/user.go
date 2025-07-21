package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func buildUserQuery(filter *api.UserFilter) string {
	params := []string{}
	showDeleted := v1pb.State_DELETED == filter.State

	if v := filter.Name; v != "" {
		params = append(params, fmt.Sprintf(`name == "%s"`, strings.ToLower(v)))
	}
	if v := filter.Email; v != "" {
		params = append(params, fmt.Sprintf(`email == "%s"`, strings.ToLower(v)))
	}
	if v := filter.Project; v != "" {
		params = append(params, fmt.Sprintf(`project == "%s"`, v))
	}
	if v := filter.UserTypes; len(v) > 0 {
		userTypes := []string{}
		for _, t := range v {
			userTypes = append(userTypes, fmt.Sprintf(`"%s"`, t.String()))
		}
		params = append(params, fmt.Sprintf(`user_type in [%s]`, strings.Join(userTypes, ", ")))
	}
	if showDeleted {
		params = append(params, fmt.Sprintf(`state == "%s"`, filter.State.String()))
	}

	if len(params) == 0 {
		return fmt.Sprintf("showDeleted=%v", showDeleted)
	}

	return fmt.Sprintf("filter=%s&showDeleted=%v", url.QueryEscape(strings.Join(params, " && ")), showDeleted)
}

// ListUser list all users.
func (c *client) ListUser(ctx context.Context, filter *api.UserFilter) ([]*v1pb.User, error) {
	res := []*v1pb.User{}
	pageToken := ""
	startTime := time.Now()
	query := buildUserQuery(filter)

	for {
		startTimePerPage := time.Now()
		resp, err := c.listUserPerPage(ctx, query, pageToken, 500)
		if err != nil {
			return nil, err
		}
		res = append(res, resp.Users...)
		tflog.Debug(ctx, "[list user per page]", map[string]interface{}{
			"count": len(resp.Users),
			"ms":    time.Since(startTimePerPage).Milliseconds(),
		})

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	tflog.Debug(ctx, "[list user]", map[string]interface{}{
		"total": len(res),
		"ms":    time.Since(startTime).Milliseconds(),
	})

	return res, nil
}

// listUserPerPage list the users.
func (c *client) listUserPerPage(ctx context.Context, query, pageToken string, pageSize int) (*v1pb.ListUsersResponse, error) {
	requestURL := fmt.Sprintf(
		"%s/%s/users?%s&page_size=%d&page_token=%s",
		c.url,
		c.version,
		query,
		pageSize,
		url.QueryEscape(pageToken),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.ListUsersResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// CreateUser creates the user.
func (c *client) CreateUser(ctx context.Context, user *v1pb.User) (*v1pb.User, error) {
	payload, err := protojson.Marshal(user)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/users", c.url, c.version), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.User
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetUser gets the user by name.
func (c *client) GetUser(ctx context.Context, userName string) (*v1pb.User, error) {
	body, err := c.getResource(ctx, userName, "")
	if err != nil {
		return nil, err
	}

	var res v1pb.User
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateUser updates the user.
func (c *client) UpdateUser(ctx context.Context, patch *v1pb.User, updateMasks []string) (*v1pb.User, error) {
	body, err := c.updateResource(ctx, patch.Name, patch, updateMasks, false /* allow missing = false*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.User
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// DeleteUser deletes the user by name.
func (c *client) DeleteUser(ctx context.Context, userName string) error {
	return c.deleteResource(ctx, userName)
}

// UndeleteUser undeletes the user by name.
func (c *client) UndeleteUser(ctx context.Context, userName string) (*v1pb.User, error) {
	body, err := c.undeleteResource(ctx, userName)
	if err != nil {
		return nil, err
	}

	var res v1pb.User
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
