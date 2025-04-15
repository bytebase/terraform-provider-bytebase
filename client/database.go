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

// GetDatabase gets the database by the database full name.
func (c *client) GetDatabase(ctx context.Context, databaseName string) (*v1pb.Database, error) {
	body, err := c.getResource(ctx, databaseName)
	if err != nil {
		return nil, err
	}

	var res v1pb.Database
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func buildDatabaseQuery(filter *api.DatabaseFilter) string {
	params := []string{}

	if v := filter.Query; v != "" {
		params = append(params, fmt.Sprintf(`name.matches("%s")`, strings.ToLower(v)))
	}
	if v := filter.Environment; v != "" {
		params = append(params, fmt.Sprintf(`environment == "%s"`, v))
	}
	if v := filter.Project; v != "" {
		params = append(params, fmt.Sprintf(`project == "%s"`, v))
	}
	if v := filter.Instance; v != "" {
		params = append(params, fmt.Sprintf(`instance == "%s"`, v))
	}
	if filter.ExcludeUnassigned {
		params = append(params, "exclude_unassigned == true")
	}
	if v := filter.Engines; len(v) > 0 {
		engines := []string{}
		for _, e := range v {
			engines = append(engines, fmt.Sprintf(`"%s"`, e.String()))
		}
		params = append(params, fmt.Sprintf(`engine in [%s]`, strings.Join(engines, ", ")))
	}
	if v := filter.Labels; len(v) > 0 {
		labelMap := map[string][]string{}
		for _, label := range v {
			if _, ok := labelMap[label.Key]; !ok {
				labelMap[label.Key] = []string{}
			}
			labelMap[label.Key] = append(labelMap[label.Key], label.Value)
		}
		for key, values := range labelMap {
			params = append(params, fmt.Sprintf(`label == "%s:%s"`, key, strings.Join(values, ",")))
		}
	}

	return fmt.Sprintf("filter=%s", url.QueryEscape(strings.Join(params, " && ")))
}

// ListDatabase list all databases.
func (c *client) ListDatabase(ctx context.Context, parent string, filter *api.DatabaseFilter, listAll bool) ([]*v1pb.Database, error) {
	res := []*v1pb.Database{}
	pageToken := ""
	startTime := time.Now()
	query := buildDatabaseQuery(filter)

	for {
		startTimePerPage := time.Now()
		resp, err := c.listDatabasePerPage(ctx, parent, query, pageToken, 500)
		if err != nil {
			return nil, err
		}
		res = append(res, resp.Databases...)
		tflog.Debug(ctx, "[list database per page]", map[string]interface{}{
			"count": len(resp.Databases),
			"ms":    time.Since(startTimePerPage).Milliseconds(),
		})

		pageToken = resp.NextPageToken
		if pageToken == "" || !listAll {
			break
		}
	}

	tflog.Debug(ctx, "[list database]", map[string]interface{}{
		"total": len(res),
		"ms":    time.Since(startTime).Milliseconds(),
	})

	return res, nil
}

// listDatabasePerPage list the databases.
func (c *client) listDatabasePerPage(ctx context.Context, parent, filter, pageToken string, pageSize int) (*v1pb.ListDatabasesResponse, error) {
	requestURL := fmt.Sprintf(
		"%s/%s/%s/databases?%s&page_size=%d&page_token=%s",
		c.url,
		c.version,
		parent,
		filter,
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

	var res v1pb.ListDatabasesResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateDatabase patches the database.
func (c *client) UpdateDatabase(ctx context.Context, patch *v1pb.Database, updateMasks []string) (*v1pb.Database, error) {
	body, err := c.updateResource(ctx, patch.Name, patch, updateMasks, false /* allow missing = false*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.Database
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// BatchUpdateDatabases batch updates databases.
func (c *client) BatchUpdateDatabases(ctx context.Context, request *v1pb.BatchUpdateDatabasesRequest) (*v1pb.BatchUpdateDatabasesResponse, error) {
	requestURL := fmt.Sprintf("%s/%s/instances/-/databases:batchUpdate", c.url, c.version)
	payload, err := protojson.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.BatchUpdateDatabasesResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetDatabaseCatalog gets the database catalog by the database full name.
func (c *client) GetDatabaseCatalog(ctx context.Context, databaseName string) (*v1pb.DatabaseCatalog, error) {
	body, err := c.getResource(ctx, fmt.Sprintf("%s/catalog", databaseName))
	if err != nil {
		return nil, err
	}

	var res v1pb.DatabaseCatalog
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateDatabaseCatalog patches the database catalog.
func (c *client) UpdateDatabaseCatalog(ctx context.Context, patch *v1pb.DatabaseCatalog) (*v1pb.DatabaseCatalog, error) {
	body, err := c.updateResource(ctx, patch.Name, patch, nil, false /* allow missing = false*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.DatabaseCatalog
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
