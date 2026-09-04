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

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func TestCreateInstanceSendsProjectParent(t *testing.T) {
	handler := &recordingInstanceHandler{}
	apiClient := newInstanceTestClient(t, handler)

	instance, err := apiClient.CreateInstance(
		context.Background(),
		"projects/sample-project",
		"project-instance",
		&v1pb.Instance{Title: "Project instance"},
	)
	if err != nil {
		t.Fatalf("CreateInstance() returned error: %v", err)
	}
	if got, want := handler.createRequest.GetParent(), "projects/sample-project"; got != want {
		t.Fatalf("CreateInstanceRequest.Parent = %q, want %q", got, want)
	}
	if got, want := instance.Name, "projects/sample-project/instances/project-instance"; got != want {
		t.Fatalf("CreateInstance() name = %q, want %q", got, want)
	}
}

func TestListInstanceSendsParentAndProjectFilter(t *testing.T) {
	handler := &recordingInstanceHandler{}
	apiClient := newInstanceTestClient(t, handler)

	_, err := apiClient.ListInstance(context.Background(), &api.InstanceFilter{
		Parent:  "projects/sample-project",
		Project: "projects/sample-project",
	})
	if err != nil {
		t.Fatalf("ListInstance() returned error: %v", err)
	}
	if got, want := handler.listRequest.GetParent(), "projects/sample-project"; got != want {
		t.Fatalf("ListInstancesRequest.Parent = %q, want %q", got, want)
	}
	if got := handler.listRequest.GetFilter(); !strings.Contains(got, `project == "projects/sample-project"`) {
		t.Fatalf("ListInstancesRequest.Filter = %q, want project filter", got)
	}
}

func TestWorkspaceInstanceRequestsOmitParent(t *testing.T) {
	handler := &recordingInstanceHandler{}
	apiClient := newInstanceTestClient(t, handler)

	if _, err := apiClient.CreateInstance(context.Background(), "", "workspace-instance", &v1pb.Instance{}); err != nil {
		t.Fatalf("CreateInstance() returned error: %v", err)
	}
	if handler.createRequest.HasParent() {
		t.Fatal("CreateInstanceRequest.Parent is set for a workspace-owned instance")
	}

	if _, err := apiClient.ListInstance(context.Background(), &api.InstanceFilter{}); err != nil {
		t.Fatalf("ListInstance() returned error: %v", err)
	}
	if handler.listRequest.HasParent() {
		t.Fatal("ListInstancesRequest.Parent is set for the workspace collection")
	}
}

func newInstanceTestClient(t *testing.T, handler bytebasev1connect.InstanceServiceHandler) *client {
	t.Helper()
	mux := http.NewServeMux()
	path, httpHandler := bytebasev1connect.NewInstanceServiceHandler(handler)
	mux.Handle(path, httpHandler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return &client{
		client:         server.Client(),
		url:            server.URL,
		instanceClient: bytebasev1connect.NewInstanceServiceClient(server.Client(), server.URL),
	}
}

type recordingInstanceHandler struct {
	bytebasev1connect.UnimplementedInstanceServiceHandler
	createRequest *v1pb.CreateInstanceRequest
	listRequest   *v1pb.ListInstancesRequest
}

func (h *recordingInstanceHandler) CreateInstance(_ context.Context, req *connect.Request[v1pb.CreateInstanceRequest]) (*connect.Response[v1pb.Instance], error) {
	h.createRequest = req.Msg
	name := "instances/" + req.Msg.GetInstanceId()
	if req.Msg.HasParent() {
		name = req.Msg.GetParent() + "/instances/" + req.Msg.GetInstanceId()
	}
	return connect.NewResponse(&v1pb.Instance{
		Name:  name,
		Title: req.Msg.GetInstance().GetTitle(),
	}), nil
}

func (h *recordingInstanceHandler) ListInstances(_ context.Context, req *connect.Request[v1pb.ListInstancesRequest]) (*connect.Response[v1pb.ListInstancesResponse], error) {
	h.listRequest = req.Msg
	return connect.NewResponse(&v1pb.ListInstancesResponse{}), nil
}
