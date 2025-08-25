package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// GetDatabase gets the database by the database full name using Connect RPC.
func (c *client) GetDatabase(ctx context.Context, databaseName string) (*v1pb.Database, error) {
	if c.databaseClient == nil {
		return nil, fmt.Errorf("database service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetDatabaseRequest{
		Name: databaseName,
	})

	resp, err := c.databaseClient.GetDatabase(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

func buildDatabaseFilter(filter *api.DatabaseFilter) string {
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

	return strings.Join(params, " && ")
}

// ListDatabase list all databases using Connect RPC.
func (c *client) ListDatabase(ctx context.Context, parent string, filter *api.DatabaseFilter, listAll bool) ([]*v1pb.Database, error) {
	if c.databaseClient == nil {
		return nil, fmt.Errorf("database service client not initialized")
	}

	res := []*v1pb.Database{}
	pageToken := ""
	startTime := time.Now()
	filterStr := buildDatabaseFilter(filter)

	for {
		startTimePerPage := time.Now()

		req := connect.NewRequest(&v1pb.ListDatabasesRequest{
			Parent:    parent,
			Filter:    filterStr,
			PageSize:  500,
			PageToken: pageToken,
		})

		resp, err := c.databaseClient.ListDatabases(ctx, req)
		if err != nil {
			return nil, err
		}

		res = append(res, resp.Msg.Databases...)

		tflog.Debug(ctx, "[list database per page]", map[string]interface{}{
			"count": len(resp.Msg.Databases),
			"ms":    time.Since(startTimePerPage).Milliseconds(),
		})

		pageToken = resp.Msg.NextPageToken
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

// UpdateDatabase patches the database using Connect RPC.
func (c *client) UpdateDatabase(ctx context.Context, patch *v1pb.Database, updateMasks []string) (*v1pb.Database, error) {
	if c.databaseClient == nil {
		return nil, fmt.Errorf("database service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdateDatabaseRequest{
		Database:   patch,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: updateMasks},
	})

	resp, err := c.databaseClient.UpdateDatabase(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// BatchUpdateDatabases batch updates databases using Connect RPC.
func (c *client) BatchUpdateDatabases(ctx context.Context, request *v1pb.BatchUpdateDatabasesRequest) (*v1pb.BatchUpdateDatabasesResponse, error) {
	if c.databaseClient == nil {
		return nil, fmt.Errorf("database service client not initialized")
	}

	req := connect.NewRequest(request)

	resp, err := c.databaseClient.BatchUpdateDatabases(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// GetDatabaseCatalog gets the database catalog by the database full name using Connect RPC.
func (c *client) GetDatabaseCatalog(ctx context.Context, databaseName string) (*v1pb.DatabaseCatalog, error) {
	if c.databaseCatalogClient == nil {
		return nil, fmt.Errorf("database catalog service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetDatabaseCatalogRequest{
		Name: fmt.Sprintf("%s/catalog", databaseName),
	})

	resp, err := c.databaseCatalogClient.GetDatabaseCatalog(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UpdateDatabaseCatalog patches the database catalog using Connect RPC.
func (c *client) UpdateDatabaseCatalog(ctx context.Context, patch *v1pb.DatabaseCatalog) (*v1pb.DatabaseCatalog, error) {
	if c.databaseCatalogClient == nil {
		return nil, fmt.Errorf("database catalog service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdateDatabaseCatalogRequest{
		Catalog: patch,
	})

	resp, err := c.databaseCatalogClient.UpdateDatabaseCatalog(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}
