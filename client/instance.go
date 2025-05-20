package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func buildInstanceQuery(filter *api.InstanceFilter) string {
	params := []string{}
	showDeleted := v1pb.State_DELETED == filter.State

	if v := filter.Query; v != "" {
		params = append(params, fmt.Sprintf(`(name.matches("%s") || resource_id.matches("%s"))`, strings.ToLower(v), strings.ToLower(v)))
	}
	if v := filter.Project; v != "" {
		params = append(params, fmt.Sprintf(`project == "%s"`, v))
	}
	if v := filter.Environment; v != "" {
		params = append(params, fmt.Sprintf(`environment == "%s"`, v))
	}
	if v := filter.Host; v != "" {
		params = append(params, fmt.Sprintf(`host == "%s"`, v))
	}
	if v := filter.Port; v != "" {
		params = append(params, fmt.Sprintf(`port == "%s"`, v))
	}
	if v := filter.Engines; len(v) > 0 {
		engines := []string{}
		for _, e := range v {
			engines = append(engines, fmt.Sprintf(`"%s"`, e.String()))
		}
		params = append(params, fmt.Sprintf(`engine in [%s]`, strings.Join(engines, ", ")))
	}
	if showDeleted {
		params = append(params, fmt.Sprintf(`state == "%s"`, filter.State.String()))
	}

	if len(params) == 0 {
		return fmt.Sprintf("showDeleted=%v", showDeleted)
	}

	return fmt.Sprintf("filter=%s&showDeleted=%v", url.QueryEscape(strings.Join(params, " && ")), showDeleted)
}

// ListInstance will return instances.
func (c *client) ListInstance(ctx context.Context, filter *api.InstanceFilter) ([]*v1pb.Instance, error) {
	res := []*v1pb.Instance{}
	pageToken := ""
	startTime := time.Now()
	query := buildInstanceQuery(filter)

	for {
		startTimePerPage := time.Now()
		resp, err := c.listInstancePerPage(ctx, query, pageToken, 500)
		if err != nil {
			return nil, err
		}
		res = append(res, resp.Instances...)
		tflog.Debug(ctx, "[list instance per page]", map[string]interface{}{
			"count": len(resp.Instances),
			"ms":    time.Since(startTimePerPage).Milliseconds(),
		})

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	tflog.Debug(ctx, "[list instance]", map[string]interface{}{
		"total": len(res),
		"ms":    time.Since(startTime).Milliseconds(),
	})

	return res, nil
}

// listInstancePerPage list the instance.
func (c *client) listInstancePerPage(ctx context.Context, query, pageToken string, pageSize int) (*v1pb.ListInstancesResponse, error) {
	requestURL := fmt.Sprintf(
		"%s/%s/instances?%s&page_size=%d&page_token=%s",
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

	var res v1pb.ListInstancesResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetInstance gets the instance by full name.
func (c *client) GetInstance(ctx context.Context, instanceName string) (*v1pb.Instance, error) {
	body, err := c.getResource(ctx, instanceName, "")
	if err != nil {
		return nil, err
	}

	var res v1pb.Instance
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// CreateInstance creates the instance.
func (c *client) CreateInstance(ctx context.Context, instanceID string, instance *v1pb.Instance) (*v1pb.Instance, error) {
	payload, err := protojson.Marshal(instance)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/instances?instanceId=%s", c.url, c.version, instanceID), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.Instance
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateInstance updates the instance.
func (c *client) UpdateInstance(ctx context.Context, patch *v1pb.Instance, updateMasks []string) (*v1pb.Instance, error) {
	body, err := c.updateResource(ctx, patch.Name, patch, updateMasks, false /* allow missing = false*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.Instance
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// DeleteInstance deletes the instance.
func (c *client) DeleteInstance(ctx context.Context, instanceName string) error {
	return c.deleteResource(ctx, instanceName)
}

// UndeleteInstance undeletes the instance.
func (c *client) UndeleteInstance(ctx context.Context, instanceName string) (*v1pb.Instance, error) {
	body, err := c.undeleteResource(ctx, instanceName)
	if err != nil {
		return nil, err
	}

	var res v1pb.Instance
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SyncInstanceSchema will trigger the schema sync for an instance.
func (c *client) SyncInstanceSchema(ctx context.Context, instanceName string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/%s:sync", c.url, c.version, instanceName), nil)

	if err != nil {
		return err
	}

	if _, err := c.doRequest(req); err != nil {
		return err
	}

	return nil
}
