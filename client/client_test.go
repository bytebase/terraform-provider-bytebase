package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"buf.build/gen/go/bytebase/bytebase/connectrpc/go/v1/bytebasev1connect"
	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
)

func TestNewClientSendsCustomHeadersToLoginAndAuthenticatedRequests(t *testing.T) {
	authHandler := &recordingAuthHandler{}
	actuatorHandler := &recordingActuatorHandler{defaultProject: "projects/default-test"}
	workspaceHandler := &recordingWorkspaceHandler{}

	mux := http.NewServeMux()
	authPath, authHTTPHandler := bytebasev1connect.NewAuthServiceHandler(authHandler)
	actuatorPath, actuatorHTTPHandler := bytebasev1connect.NewActuatorServiceHandler(actuatorHandler)
	workspacePath, workspaceHTTPHandler := bytebasev1connect.NewWorkspaceServiceHandler(workspaceHandler)
	mux.Handle(authPath, authHTTPHandler)
	mux.Handle(actuatorPath, actuatorHTTPHandler)
	mux.Handle(workspacePath, workspaceHTTPHandler)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	headers := map[string]string{
		"zero_trust_token": "test-zero-trust-token",
		"X-Bytebase-Test":  "test-value",
	}
	apiClient, err := NewClient(server.URL, "service@example.com", "secret", WithCustomHeaders(headers))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if got, want := apiClient.GetDefaultProjectName(), "projects/default-test"; got != want {
		t.Fatalf("GetDefaultProjectName() = %q, want %q", got, want)
	}
	if _, err := apiClient.GetWorkspace(context.Background(), "workspaces/test"); err != nil {
		t.Fatalf("GetWorkspace() error = %v", err)
	}

	for name, value := range headers {
		if got := authHandler.headers.Get(name); got != value {
			t.Fatalf("login header %q = %q, want %q", name, got, value)
		}
		if got := actuatorHandler.headers.Get(name); got != value {
			t.Fatalf("actuator header %q = %q, want %q", name, got, value)
		}
		if got := workspaceHandler.headers.Get(name); got != value {
			t.Fatalf("workspace header %q = %q, want %q", name, got, value)
		}
	}
	if got := actuatorHandler.headers.Get("Authorization"); got != "Bearer test-token" {
		t.Fatalf("actuator Authorization header = %q, want %q", got, "Bearer test-token")
	}
	if got := workspaceHandler.headers.Get("Authorization"); got != "Bearer test-token" {
		t.Fatalf("workspace Authorization header = %q, want %q", got, "Bearer test-token")
	}
}

func TestNewClientRequiresDefaultProject(t *testing.T) {
	authHandler := &recordingAuthHandler{}
	actuatorHandler := &recordingActuatorHandler{}

	mux := http.NewServeMux()
	authPath, authHTTPHandler := bytebasev1connect.NewAuthServiceHandler(authHandler)
	actuatorPath, actuatorHTTPHandler := bytebasev1connect.NewActuatorServiceHandler(actuatorHandler)
	mux.Handle(authPath, authHTTPHandler)
	mux.Handle(actuatorPath, actuatorHTTPHandler)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	_, err := NewClient(server.URL, "service@example.com", "secret")
	if err == nil {
		t.Fatal("NewClient() error = nil, want error")
	}
	if got, want := err.Error(), "actuator returned empty default project name"; !strings.Contains(got, want) {
		t.Fatalf("NewClient() error = %q, want to contain %q", got, want)
	}
}

type recordingAuthHandler struct {
	bytebasev1connect.UnimplementedAuthServiceHandler
	headers http.Header
}

func (h *recordingAuthHandler) Login(_ context.Context, req *connect.Request[v1pb.LoginRequest]) (*connect.Response[v1pb.LoginResponse], error) {
	h.headers = req.Header()
	return connect.NewResponse(&v1pb.LoginResponse{Token: "test-token"}), nil
}

type recordingActuatorHandler struct {
	bytebasev1connect.UnimplementedActuatorServiceHandler
	headers        http.Header
	defaultProject string
}

func (h *recordingActuatorHandler) GetActuatorInfo(_ context.Context, req *connect.Request[v1pb.GetActuatorInfoRequest]) (*connect.Response[v1pb.ActuatorInfo], error) {
	h.headers = req.Header()
	return connect.NewResponse(&v1pb.ActuatorInfo{
		Workspace:      "workspaces/test",
		DefaultProject: h.defaultProject,
	}), nil
}

type recordingWorkspaceHandler struct {
	bytebasev1connect.UnimplementedWorkspaceServiceHandler
	headers http.Header
}

func (h *recordingWorkspaceHandler) GetWorkspace(_ context.Context, req *connect.Request[v1pb.GetWorkspaceRequest]) (*connect.Response[v1pb.Workspace], error) {
	h.headers = req.Header()
	return connect.NewResponse(&v1pb.Workspace{Name: "workspaces/test"}), nil
}
