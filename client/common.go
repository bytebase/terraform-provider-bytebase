package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ProtojsonUnmarshaler is the unmarshal for protocol.
var ProtojsonUnmarshaler = protojson.UnmarshalOptions{DiscardUnknown: true}

// deleteResource deletes the resource by name.
func (c *client) deleteResource(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/%s/%s?force=true", c.url, c.version, url.QueryEscape(name)), nil)
	if err != nil {
		return err
	}

	if _, err := c.doRequest(req); err != nil {
		return err
	}
	return nil
}

// undeleteResource undeletes the resource by name.
func (c *client) undeleteResource(ctx context.Context, name string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/%s:undelete", c.url, c.version, url.QueryEscape(name)), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// deleteResource deletes the resource by name.
func (c *client) updateResource(ctx context.Context, name string, patch protoreflect.ProtoMessage, updateMasks []string, allowMissing bool) ([]byte, error) {
	payload, err := protojson.Marshal(patch)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/%s/%s?update_mask=%s&allow_missing=%v", c.url, c.version, url.QueryEscape(name), strings.Join(updateMasks, ","), allowMissing), strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// getResource gets the resource by name.
func (c *client) getResource(ctx context.Context, name, query string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/%s?%s", c.url, c.version, url.QueryEscape(name), query), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	return body, nil
}
