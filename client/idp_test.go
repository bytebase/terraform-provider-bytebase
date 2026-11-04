package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"buf.build/gen/go/bytebase/bytebase/connectrpc/go/v1/bytebasev1connect"
	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
)

func TestListIdentityProviderSendsWorkspaceParent(t *testing.T) {
	mux := http.NewServeMux()
	path, handler := bytebasev1connect.NewIdentityProviderServiceHandler(&workspaceIdentityProviderHandler{})
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	apiClient := &client{
		workspaceName: "workspaces/test-workspace",
		idpClient:     bytebasev1connect.NewIdentityProviderServiceClient(server.Client(), server.URL),
	}

	providers, err := apiClient.ListIdentityProvider(context.Background())
	if err != nil {
		t.Fatalf("ListIdentityProvider() returned error: %v", err)
	}
	if len(providers) != 1 || providers[0].Name != "idps/test-idp" {
		t.Fatalf("ListIdentityProvider() = %v, want the workspace identity provider", providers)
	}
}

type workspaceIdentityProviderHandler struct {
	bytebasev1connect.UnimplementedIdentityProviderServiceHandler
}

func (*workspaceIdentityProviderHandler) ListIdentityProviders(_ context.Context, request *connect.Request[v1pb.ListIdentityProvidersRequest]) (*connect.Response[v1pb.ListIdentityProvidersResponse], error) {
	if request.Msg.GetParent() != "workspaces/test-workspace" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("parent must match the credential workspace"))
	}
	return connect.NewResponse(&v1pb.ListIdentityProvidersResponse{
		IdentityProviders: []*v1pb.IdentityProvider{{Name: "idps/test-idp"}},
	}), nil
}
