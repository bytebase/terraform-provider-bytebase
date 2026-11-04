package client

import (
	"context"
	"errors"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ListIdentityProvider lists all identity providers.
func (c *client) ListIdentityProvider(ctx context.Context) ([]*v1pb.IdentityProvider, error) {
	if c.idpClient == nil {
		return nil, errors.New("identity provider service client not initialized")
	}

	req := connect.NewRequest(&v1pb.ListIdentityProvidersRequest{})

	resp, err := c.idpClient.ListIdentityProviders(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg.IdentityProviders, nil
}

// GetIdentityProvider gets the identity provider by name.
func (c *client) GetIdentityProvider(ctx context.Context, name string) (*v1pb.IdentityProvider, error) {
	if c.idpClient == nil {
		return nil, errors.New("identity provider service client not initialized")
	}

	req := connect.NewRequest(&v1pb.GetIdentityProviderRequest{
		Name: name,
	})

	resp, err := c.idpClient.GetIdentityProvider(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// CreateIdentityProvider creates the identity provider.
func (c *client) CreateIdentityProvider(ctx context.Context, idpID string, idp *v1pb.IdentityProvider) (*v1pb.IdentityProvider, error) {
	if c.idpClient == nil {
		return nil, errors.New("identity provider service client not initialized")
	}

	req := connect.NewRequest(&v1pb.CreateIdentityProviderRequest{
		IdentityProvider:   idp,
		IdentityProviderId: idpID,
	})

	resp, err := c.idpClient.CreateIdentityProvider(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UpdateIdentityProvider updates the identity provider.
func (c *client) UpdateIdentityProvider(ctx context.Context, patch *v1pb.IdentityProvider, updateMasks []string) (*v1pb.IdentityProvider, error) {
	if c.idpClient == nil {
		return nil, errors.New("identity provider service client not initialized")
	}

	req := connect.NewRequest(&v1pb.UpdateIdentityProviderRequest{
		IdentityProvider: patch,
		UpdateMask:       &fieldmaskpb.FieldMask{Paths: updateMasks},
	})

	resp, err := c.idpClient.UpdateIdentityProvider(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// DeleteIdentityProvider deletes the identity provider.
func (c *client) DeleteIdentityProvider(ctx context.Context, name string) error {
	if c.idpClient == nil {
		return errors.New("identity provider service client not initialized")
	}

	req := connect.NewRequest(&v1pb.DeleteIdentityProviderRequest{
		Name: name,
	})

	_, err := c.idpClient.DeleteIdentityProvider(ctx, req)
	return err
}
