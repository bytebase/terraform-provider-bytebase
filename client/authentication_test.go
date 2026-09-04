package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"buf.build/gen/go/bytebase/bytebase/connectrpc/go/v1/bytebasev1connect"
	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
)

func TestInlineTokenSource(t *testing.T) {
	source := &inlineTokenSource{token: "external-token"}
	token, err := source.Token(context.Background())
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if token != "external-token" {
		t.Fatalf("Token() = %q, want %q", token, "external-token")
	}
}

func TestFileTokenSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "workload-token.jwt")
	if err := os.WriteFile(path, []byte("first-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	source := &fileTokenSource{path: path}

	token, err := source.Token(context.Background())
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if token != "first-token" {
		t.Fatalf("Token() = %q, want %q", token, "first-token")
	}

	replacement := filepath.Join(filepath.Dir(path), "replacement.jwt")
	if err := os.WriteFile(replacement, []byte(" second-token \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, path); err != nil {
		t.Fatal(err)
	}
	token, err = source.Token(context.Background())
	if err != nil {
		t.Fatalf("Token() after replacement error = %v", err)
	}
	if token != "second-token" {
		t.Fatalf("Token() after replacement = %q, want %q", token, "second-token")
	}

	if err := os.WriteFile(path, []byte(" \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := source.Token(context.Background()); err == nil {
		t.Fatal("Token() with empty file error = nil, want error")
	}
}

func TestFileTokenSourceErrorDoesNotExposeContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.jwt")
	source := &fileTokenSource{path: path}
	_, err := source.Token(context.Background())
	if err == nil {
		t.Fatal("Token() error = nil, want error")
	}
	if !strings.Contains(err.Error(), path) {
		t.Fatalf("Token() error = %q, want path %q", err, path)
	}
}

func TestNewClientWithWorkloadIdentity(t *testing.T) {
	authHandler := &authenticationTestAuthHandler{accessTokens: []string{"bytebase-token"}}
	actuatorHandler := &recordingActuatorHandler{defaultProject: "projects/default-test"}

	mux := http.NewServeMux()
	authPath, authHTTPHandler := bytebasev1connect.NewAuthServiceHandler(authHandler)
	actuatorPath, actuatorHTTPHandler := bytebasev1connect.NewActuatorServiceHandler(actuatorHandler)
	mux.Handle(authPath, authHTTPHandler)
	mux.Handle(actuatorPath, actuatorHTTPHandler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	apiClient, err := NewClientWithAuthentication(server.URL, AuthenticationConfig{
		WorkloadIdentity: &WorkloadIdentityAuthentication{
			Email: "terraform@workload.bytebase.com",
			Token: "external-token",
		},
	}, WithCustomHeaders(map[string]string{"X-Test-Header": "header-value"}))
	if err != nil {
		t.Fatalf("NewClientWithAuthentication() error = %v", err)
	}
	if apiClient.GetDefaultProjectName() != "projects/default-test" {
		t.Fatalf("GetDefaultProjectName() = %q", apiClient.GetDefaultProjectName())
	}
	if authHandler.exchangeEmails[0] != "terraform@workload.bytebase.com" {
		t.Fatalf("exchange email = %q", authHandler.exchangeEmails[0])
	}
	if authHandler.exchangeTokens[0] != "external-token" {
		t.Fatalf("exchange token = %q", authHandler.exchangeTokens[0])
	}
	if authHandler.headers[0].Get("X-Test-Header") != "header-value" {
		t.Fatalf("exchange custom header = %q", authHandler.headers[0].Get("X-Test-Header"))
	}
	if actuatorHandler.headers.Get("Authorization") != "Bearer bytebase-token" {
		t.Fatalf("actuator Authorization = %q", actuatorHandler.headers.Get("Authorization"))
	}
}

func TestWorkloadIdentityAuthenticationErrorDoesNotExposeToken(t *testing.T) {
	const externalToken = "external-token-must-not-leak"
	authHandler := &authenticationTestAuthHandler{exchangeErr: connect.NewError(connect.CodeUnauthenticated, errors.New(externalToken))}
	mux := http.NewServeMux()
	authPath, authHTTPHandler := bytebasev1connect.NewAuthServiceHandler(authHandler)
	mux.Handle(authPath, authHTTPHandler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	_, err := NewClientWithAuthentication(server.URL, AuthenticationConfig{
		WorkloadIdentity: &WorkloadIdentityAuthentication{
			Email: "terraform@workload.bytebase.com",
			Token: externalToken,
		},
	})
	if err == nil {
		t.Fatal("NewClientWithAuthentication() error = nil, want error")
	}
	if strings.Contains(err.Error(), externalToken) {
		t.Fatalf("NewClientWithAuthentication() error exposes token: %q", err)
	}
}

func TestWorkloadIdentityRefreshesAndRetriesOnce(t *testing.T) {
	authHandler := &authenticationTestAuthHandler{accessTokens: []string{"initial-token", "refreshed-token"}}
	actuatorHandler := &recordingActuatorHandler{defaultProject: "projects/default-test"}
	workspaceHandler := newRefreshWorkspaceHandler(1)
	server := newAuthenticationTestServer(t, authHandler, actuatorHandler, workspaceHandler)

	apiClient, err := NewClientWithAuthentication(server.URL, AuthenticationConfig{
		WorkloadIdentity: &WorkloadIdentityAuthentication{
			Email: "terraform@workload.bytebase.com",
			Token: "external-token",
		},
	}, WithCustomHeaders(map[string]string{"X-Test-Header": "header-value"}))
	if err != nil {
		t.Fatalf("NewClientWithAuthentication() error = %v", err)
	}
	if _, err := apiClient.GetWorkspace(context.Background(), "workspaces/test"); err != nil {
		t.Fatalf("GetWorkspace() error = %v", err)
	}

	if got := authHandler.exchangeCount(); got != 2 {
		t.Fatalf("token exchanges = %d, want 2", got)
	}
	for index, headers := range authHandler.headerSnapshot() {
		if got := headers.Get("X-Test-Header"); got != "header-value" {
			t.Fatalf("exchange %d custom header = %q, want %q", index, got, "header-value")
		}
	}
	initial, refreshed := workspaceHandler.requestCounts()
	if initial != 1 || refreshed != 1 {
		t.Fatalf("workspace requests with initial/refreshed token = %d/%d, want 1/1", initial, refreshed)
	}
}

func TestServiceAccountRefreshesAndRetriesOnce(t *testing.T) {
	authHandler := &authenticationTestAuthHandler{loginTokens: []string{"initial-token", "refreshed-token"}}
	actuatorHandler := &recordingActuatorHandler{defaultProject: "projects/default-test"}
	workspaceHandler := newRefreshWorkspaceHandler(1)
	server := newAuthenticationTestServer(t, authHandler, actuatorHandler, workspaceHandler)

	apiClient, err := NewClient(server.URL, "service@example.com", "service-key")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if _, err := apiClient.GetWorkspace(context.Background(), "workspaces/test"); err != nil {
		t.Fatalf("GetWorkspace() error = %v", err)
	}
	if got := authHandler.loginCount(); got != 2 {
		t.Fatalf("service account logins = %d, want 2", got)
	}
}

func TestWorkloadIdentityParallelFailuresShareRefresh(t *testing.T) {
	const parallelRequests = 8
	tokenPath := filepath.Join(t.TempDir(), "workload-token.jwt")
	if err := os.WriteFile(tokenPath, []byte("external-token-one"), 0o600); err != nil {
		t.Fatal(err)
	}

	authHandler := &authenticationTestAuthHandler{accessTokens: []string{"initial-token", "refreshed-token"}}
	actuatorHandler := &recordingActuatorHandler{defaultProject: "projects/default-test"}
	workspaceHandler := newRefreshWorkspaceHandler(parallelRequests)
	server := newAuthenticationTestServer(t, authHandler, actuatorHandler, workspaceHandler)

	apiClient, err := NewClientWithAuthentication(server.URL, AuthenticationConfig{
		WorkloadIdentity: &WorkloadIdentityAuthentication{
			Email:     "terraform@workload.bytebase.com",
			TokenFile: tokenPath,
		},
	})
	if err != nil {
		t.Fatalf("NewClientWithAuthentication() error = %v", err)
	}
	replacement := filepath.Join(filepath.Dir(tokenPath), "replacement.jwt")
	if err := os.WriteFile(replacement, []byte("external-token-two\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, tokenPath); err != nil {
		t.Fatal(err)
	}

	var waitGroup sync.WaitGroup
	errorsByRequest := make(chan error, parallelRequests)
	for range parallelRequests {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, err := apiClient.GetWorkspace(context.Background(), "workspaces/test")
			errorsByRequest <- err
		}()
	}
	waitGroup.Wait()
	close(errorsByRequest)
	for err := range errorsByRequest {
		if err != nil {
			t.Fatalf("GetWorkspace() error = %v", err)
		}
	}

	if got := authHandler.exchangeCount(); got != 2 {
		t.Fatalf("token exchanges = %d, want initial exchange plus one shared refresh", got)
	}
	if got := authHandler.exchangeTokenSnapshot(); len(got) != 2 || got[0] != "external-token-one" || got[1] != "external-token-two" {
		t.Fatalf("external tokens exchanged = %v, want rotated file content", got)
	}
	initial, refreshed := workspaceHandler.requestCounts()
	if initial != parallelRequests || refreshed != parallelRequests {
		t.Fatalf("workspace requests with initial/refreshed token = %d/%d, want %d/%d", initial, refreshed, parallelRequests, parallelRequests)
	}
}

func TestWorkloadIdentityRetryLimit(t *testing.T) {
	authHandler := &authenticationTestAuthHandler{accessTokens: []string{"initial-token", "refreshed-token", "unexpected-token"}}
	actuatorHandler := &recordingActuatorHandler{defaultProject: "projects/default-test"}
	workspaceHandler := &failingWorkspaceHandler{code: connect.CodeUnauthenticated}
	server := newAuthenticationTestServer(t, authHandler, actuatorHandler, workspaceHandler)

	apiClient, err := NewClientWithAuthentication(server.URL, AuthenticationConfig{
		WorkloadIdentity: &WorkloadIdentityAuthentication{Email: "terraform@workload.bytebase.com", Token: "external-token"},
	})
	if err != nil {
		t.Fatalf("NewClientWithAuthentication() error = %v", err)
	}
	if _, err := apiClient.GetWorkspace(context.Background(), "workspaces/test"); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("GetWorkspace() code = %v, want %v", connect.CodeOf(err), connect.CodeUnauthenticated)
	}
	if got := authHandler.exchangeCount(); got != 2 {
		t.Fatalf("token exchanges = %d, want 2", got)
	}
	if got := workspaceHandler.requestCount(); got != 2 {
		t.Fatalf("workspace requests = %d, want 2", got)
	}
}

func TestNonAuthenticationErrorIsNotRetried(t *testing.T) {
	authHandler := &authenticationTestAuthHandler{accessTokens: []string{"initial-token"}}
	actuatorHandler := &recordingActuatorHandler{defaultProject: "projects/default-test"}
	workspaceHandler := &failingWorkspaceHandler{code: connect.CodePermissionDenied}
	server := newAuthenticationTestServer(t, authHandler, actuatorHandler, workspaceHandler)

	apiClient, err := NewClientWithAuthentication(server.URL, AuthenticationConfig{
		WorkloadIdentity: &WorkloadIdentityAuthentication{Email: "terraform@workload.bytebase.com", Token: "external-token"},
	})
	if err != nil {
		t.Fatalf("NewClientWithAuthentication() error = %v", err)
	}
	if _, err := apiClient.GetWorkspace(context.Background(), "workspaces/test"); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("GetWorkspace() code = %v, want %v", connect.CodeOf(err), connect.CodePermissionDenied)
	}
	if got := authHandler.exchangeCount(); got != 1 {
		t.Fatalf("token exchanges = %d, want 1", got)
	}
	if got := workspaceHandler.requestCount(); got != 1 {
		t.Fatalf("workspace requests = %d, want 1", got)
	}
}

func newAuthenticationTestServer(
	t *testing.T,
	authHandler bytebasev1connect.AuthServiceHandler,
	actuatorHandler bytebasev1connect.ActuatorServiceHandler,
	workspaceHandler bytebasev1connect.WorkspaceServiceHandler,
) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	authPath, authHTTPHandler := bytebasev1connect.NewAuthServiceHandler(authHandler)
	actuatorPath, actuatorHTTPHandler := bytebasev1connect.NewActuatorServiceHandler(actuatorHandler)
	workspacePath, workspaceHTTPHandler := bytebasev1connect.NewWorkspaceServiceHandler(workspaceHandler)
	mux.Handle(authPath, authHTTPHandler)
	mux.Handle(actuatorPath, actuatorHTTPHandler)
	mux.Handle(workspacePath, workspaceHTTPHandler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

type authenticationTestAuthHandler struct {
	bytebasev1connect.UnimplementedAuthServiceHandler
	mu             sync.Mutex
	accessTokens   []string
	loginTokens    []string
	exchangeErr    error
	exchangeEmails []string
	exchangeTokens []string
	headers        []http.Header
	loginHeaders   []http.Header
}

func (h *authenticationTestAuthHandler) Login(_ context.Context, req *connect.Request[v1pb.LoginRequest]) (*connect.Response[v1pb.LoginResponse], error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	index := len(h.loginHeaders)
	h.loginHeaders = append(h.loginHeaders, req.Header().Clone())
	return connect.NewResponse(&v1pb.LoginResponse{Token: h.loginTokens[index]}), nil
}

func (h *authenticationTestAuthHandler) ExchangeToken(_ context.Context, req *connect.Request[v1pb.ExchangeTokenRequest]) (*connect.Response[v1pb.ExchangeTokenResponse], error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.exchangeEmails = append(h.exchangeEmails, req.Msg.Email)
	h.exchangeTokens = append(h.exchangeTokens, req.Msg.Token)
	h.headers = append(h.headers, req.Header().Clone())
	if h.exchangeErr != nil {
		return nil, h.exchangeErr
	}
	index := len(h.exchangeTokens) - 1
	return connect.NewResponse(&v1pb.ExchangeTokenResponse{AccessToken: h.accessTokens[index]}), nil
}

func (h *authenticationTestAuthHandler) exchangeCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.exchangeTokens)
}

func (h *authenticationTestAuthHandler) exchangeTokenSnapshot() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string(nil), h.exchangeTokens...)
}

func (h *authenticationTestAuthHandler) headerSnapshot() []http.Header {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]http.Header(nil), h.headers...)
}

func (h *authenticationTestAuthHandler) loginCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.loginHeaders)
}

type refreshWorkspaceHandler struct {
	bytebasev1connect.UnimplementedWorkspaceServiceHandler
	mu                sync.Mutex
	waitForInitial    int
	initialRequests   int
	refreshedRequests int
	initialReady      chan struct{}
}

func newRefreshWorkspaceHandler(waitForInitial int) *refreshWorkspaceHandler {
	return &refreshWorkspaceHandler{waitForInitial: waitForInitial, initialReady: make(chan struct{})}
}

func (h *refreshWorkspaceHandler) GetWorkspace(_ context.Context, req *connect.Request[v1pb.GetWorkspaceRequest]) (*connect.Response[v1pb.Workspace], error) {
	switch req.Header().Get("Authorization") {
	case "Bearer initial-token":
		h.mu.Lock()
		h.initialRequests++
		if h.initialRequests == h.waitForInitial {
			close(h.initialReady)
		}
		h.mu.Unlock()
		<-h.initialReady
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("expired token"))
	case "Bearer refreshed-token":
		h.mu.Lock()
		h.refreshedRequests++
		h.mu.Unlock()
		return connect.NewResponse(&v1pb.Workspace{Name: "workspaces/test"}), nil
	default:
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unexpected token"))
	}
}

func (h *refreshWorkspaceHandler) requestCounts() (int, int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.initialRequests, h.refreshedRequests
}

type failingWorkspaceHandler struct {
	bytebasev1connect.UnimplementedWorkspaceServiceHandler
	mu       sync.Mutex
	code     connect.Code
	requests int
}

func (h *failingWorkspaceHandler) GetWorkspace(context.Context, *connect.Request[v1pb.GetWorkspaceRequest]) (*connect.Response[v1pb.Workspace], error) {
	h.mu.Lock()
	h.requests++
	h.mu.Unlock()
	return nil, connect.NewError(h.code, errors.New("request failed"))
}

func (h *failingWorkspaceHandler) requestCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.requests
}
