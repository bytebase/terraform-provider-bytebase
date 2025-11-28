// Package client is the API message for Bytebase API client.
package client

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"buf.build/gen/go/bytebase/bytebase/connectrpc/go/v1/bytebasev1connect"
	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// client is the API message for Bytebase API client.
type client struct {
	url    string
	client *http.Client

	// Connect RPC clients
	authClient            bytebasev1connect.AuthServiceClient
	workspaceClient       bytebasev1connect.WorkspaceServiceClient
	instanceClient        bytebasev1connect.InstanceServiceClient
	databaseClient        bytebasev1connect.DatabaseServiceClient
	databaseCatalogClient bytebasev1connect.DatabaseCatalogServiceClient
	databaseGroupClient   bytebasev1connect.DatabaseGroupServiceClient
	projectClient         bytebasev1connect.ProjectServiceClient
	userClient            bytebasev1connect.UserServiceClient
	roleClient            bytebasev1connect.RoleServiceClient
	groupClient           bytebasev1connect.GroupServiceClient
	settingClient         bytebasev1connect.SettingServiceClient
	orgPolicyClient       bytebasev1connect.OrgPolicyServiceClient
	reviewConfigClient bytebasev1connect.ReviewConfigServiceClient
	celClient          bytebasev1connect.CelServiceClient
}

// NewClient returns the new Bytebase API client.
func NewClient(url, email, password string) (api.Client, error) {
	c := client{
		url: strings.TrimSuffix(url, "/"),
	}

	// Use standard HTTP client that supports both HTTP/1.1 and HTTP/2
	c.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	authInt := &authInterceptor{}
	interceptors := connect.WithInterceptors(authInt)

	// Create auth client without token first
	// Try without WithGRPC first to see if it's a standard Connect/gRPC-Web service
	c.authClient = bytebasev1connect.NewAuthServiceClient(
		c.client,
		c.url,
	)

	// Login to get token
	loginReq := connect.NewRequest(&v1pb.LoginRequest{
		Email:    email,
		Password: password,
	})

	loginResp, err := c.authClient.Login(context.Background(), loginReq)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to login")
	}

	authInt.token = loginResp.Msg.Token

	// Initialize other clients with auth token
	c.workspaceClient = bytebasev1connect.NewWorkspaceServiceClient(c.client, c.url, interceptors)
	c.instanceClient = bytebasev1connect.NewInstanceServiceClient(c.client, c.url, interceptors)
	c.databaseClient = bytebasev1connect.NewDatabaseServiceClient(c.client, c.url, interceptors)
	c.databaseCatalogClient = bytebasev1connect.NewDatabaseCatalogServiceClient(c.client, c.url, interceptors)
	c.databaseGroupClient = bytebasev1connect.NewDatabaseGroupServiceClient(c.client, c.url, interceptors)
	c.projectClient = bytebasev1connect.NewProjectServiceClient(c.client, c.url, interceptors)
	c.userClient = bytebasev1connect.NewUserServiceClient(c.client, c.url, interceptors)
	c.roleClient = bytebasev1connect.NewRoleServiceClient(c.client, c.url, interceptors)
	c.groupClient = bytebasev1connect.NewGroupServiceClient(c.client, c.url, interceptors)
	c.settingClient = bytebasev1connect.NewSettingServiceClient(c.client, c.url, interceptors)
	c.orgPolicyClient = bytebasev1connect.NewOrgPolicyServiceClient(c.client, c.url, interceptors)
	c.reviewConfigClient = bytebasev1connect.NewReviewConfigServiceClient(c.client, c.url, interceptors)
	c.celClient = bytebasev1connect.NewCelServiceClient(c.client, c.url, interceptors)

	return &c, nil
}
