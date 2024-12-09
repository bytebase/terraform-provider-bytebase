package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// ListVCSProvider will returns all vcs providers.
func (c *client) ListVCSProvider(ctx context.Context) (*v1pb.ListVCSProvidersResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/vcsProviders", c.url, c.version), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.ListVCSProvidersResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetVCSProvider gets the vcs by id.
func (c *client) GetVCSProvider(ctx context.Context, name string) (*v1pb.VCSProvider, error) {
	body, err := c.getResource(ctx, name)
	if err != nil {
		return nil, err
	}

	var res v1pb.VCSProvider
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// CreateVCSProvider creates the vcs provider.
func (c *client) CreateVCSProvider(ctx context.Context, vcsID string, vcs *v1pb.VCSProvider) (*v1pb.VCSProvider, error) {
	payload, err := protojson.Marshal(vcs)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/vcsProviders?vcsProviderId=%s", c.url, c.version, vcsID), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.VCSProvider
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateVCSProvider updates the vcs provider.
func (c *client) UpdateVCSProvider(ctx context.Context, patch *v1pb.VCSProvider, updateMasks []string) (*v1pb.VCSConnector, error) {
	body, err := c.updateResource(ctx, patch.Name, patch, updateMasks, false /* allow missing = false*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.VCSConnector
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// DeleteVCSProvider deletes the vcs provider.
func (c *client) DeleteVCSProvider(ctx context.Context, name string) error {
	return c.deleteResource(ctx, name)
}

// ListVCSConnector will returns all vcs connector in a project.
func (c *client) ListVCSConnector(ctx context.Context, projectName string) (*v1pb.ListVCSConnectorsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/%s/vcsConnectors", c.url, c.version, projectName), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.ListVCSConnectorsResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetVCSConnector gets the vcs connector by id.
func (c *client) GetVCSConnector(ctx context.Context, name string) (*v1pb.VCSConnector, error) {
	body, err := c.getResource(ctx, name)
	if err != nil {
		return nil, err
	}

	var res v1pb.VCSConnector
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// CreateVCSConnector creates the vcs connector in a project.
func (c *client) CreateVCSConnector(ctx context.Context, projectName, connectorID string, connector *v1pb.VCSConnector) (*v1pb.VCSConnector, error) {
	payload, err := protojson.Marshal(connector)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/%s/vcsConnectors?vcsConnectorId=%s", c.url, c.version, projectName, connectorID), strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var res v1pb.VCSConnector
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateVCSConnector updates the vcs connector.
func (c *client) UpdateVCSConnector(ctx context.Context, patch *v1pb.VCSConnector, updateMasks []string) (*v1pb.VCSConnector, error) {
	body, err := c.updateResource(ctx, patch.Name, patch, updateMasks, false /* allow missing = false*/)
	if err != nil {
		return nil, err
	}

	var res v1pb.VCSConnector
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// DeleteVCSConnector deletes the vcs provider.
func (c *client) DeleteVCSConnector(ctx context.Context, name string) error {
	return c.deleteResource(ctx, name)
}
