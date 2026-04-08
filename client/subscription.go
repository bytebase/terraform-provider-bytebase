package client

import (
	"context"
	"errors"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
)

// GetSubscription gets the current subscription.
func (c *client) GetSubscription(ctx context.Context) (*v1pb.Subscription, error) {
	if c.subscriptionClient == nil {
		return nil, errors.New("subscription service client not initialized")
	}

	resp, err := c.subscriptionClient.GetSubscription(ctx, connect.NewRequest(&v1pb.GetSubscriptionRequest{}))
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

// UploadLicense uploads a license.
func (c *client) UploadLicense(ctx context.Context, license string) (*v1pb.Subscription, error) {
	if c.subscriptionClient == nil {
		return nil, errors.New("subscription service client not initialized")
	}

	resp, err := c.subscriptionClient.UploadLicense(ctx, connect.NewRequest(&v1pb.UploadLicenseRequest{
		License: license,
	}))
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}
