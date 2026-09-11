package client

import (
	"context"
	"os"
	"strings"

	"buf.build/gen/go/bytebase/bytebase/connectrpc/go/v1/bytebasev1connect"
	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/pkg/errors"
)

// AuthenticationConfig configures exactly one Bytebase authentication mode.
type AuthenticationConfig struct {
	ServiceAccount   *ServiceAccountAuthentication
	WorkloadIdentity *WorkloadIdentityAuthentication
}

// ServiceAccountAuthentication configures service account authentication.
type ServiceAccountAuthentication struct {
	Email string
	Key   string
}

// WorkloadIdentityAuthentication configures workload identity authentication.
type WorkloadIdentityAuthentication struct {
	Email     string
	Token     string
	TokenFile string
}

type externalTokenSource interface {
	Token(context.Context) (string, error)
}

type inlineTokenSource struct {
	token string
}

func (s *inlineTokenSource) Token(context.Context) (string, error) {
	return s.token, nil
}

type fileTokenSource struct {
	path string
}

func (s *fileTokenSource) Token(context.Context) (string, error) {
	content, err := os.ReadFile(s.path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read workload identity token file %q", s.path)
	}
	token := strings.TrimSpace(string(content))
	if token == "" {
		return "", errors.Errorf("workload identity token file %q is empty", s.path)
	}
	return token, nil
}

type authenticator interface {
	Authenticate(context.Context) (string, error)
}

type serviceAccountAuthenticator struct {
	authClient    bytebasev1connect.AuthServiceClient
	email         string
	key           string
	customHeaders map[string]string
}

func (a *serviceAccountAuthenticator) Authenticate(ctx context.Context) (string, error) {
	req := connect.NewRequest(&v1pb.LoginRequest{
		Email:    a.email,
		Password: a.key,
	})
	setHeaders(req.Header(), a.customHeaders)
	resp, err := a.authClient.Login(ctx, req)
	if err != nil {
		return "", sanitizedAuthenticationError("service account login", err)
	}
	return resp.Msg.Token, nil
}

type workloadIdentityAuthenticator struct {
	authClient    bytebasev1connect.AuthServiceClient
	email         string
	tokenSource   externalTokenSource
	customHeaders map[string]string
}

func (a *workloadIdentityAuthenticator) Authenticate(ctx context.Context) (string, error) {
	token, err := a.tokenSource.Token(ctx)
	if err != nil {
		return "", err
	}
	req := connect.NewRequest(&v1pb.ExchangeTokenRequest{
		Email: a.email,
		Token: token,
	})
	setHeaders(req.Header(), a.customHeaders)
	resp, err := a.authClient.ExchangeToken(ctx, req)
	if err != nil {
		return "", sanitizedAuthenticationError("workload identity token exchange", err)
	}
	return resp.Msg.AccessToken, nil
}

func newAuthenticator(authClient bytebasev1connect.AuthServiceClient, config AuthenticationConfig, customHeaders map[string]string) (authenticator, error) {
	if config.ServiceAccount != nil && config.WorkloadIdentity != nil {
		return nil, errors.New("service account and workload identity authentication cannot be configured together")
	}
	if service := config.ServiceAccount; service != nil {
		if service.Email == "" || service.Key == "" {
			return nil, errors.New("service account email and key are required")
		}
		return &serviceAccountAuthenticator{
			authClient:    authClient,
			email:         service.Email,
			key:           service.Key,
			customHeaders: copyHeaders(customHeaders),
		}, nil
	}
	if workload := config.WorkloadIdentity; workload != nil {
		if workload.Email == "" {
			return nil, errors.New("workload identity email is required")
		}
		if (workload.Token == "") == (workload.TokenFile == "") {
			return nil, errors.New("exactly one workload identity token source is required")
		}
		var source externalTokenSource = &inlineTokenSource{token: workload.Token}
		if workload.TokenFile != "" {
			source = &fileTokenSource{path: workload.TokenFile}
		}
		return &workloadIdentityAuthenticator{
			authClient:    authClient,
			email:         workload.Email,
			tokenSource:   source,
			customHeaders: copyHeaders(customHeaders),
		}, nil
	}
	return nil, errors.New("authentication configuration is required")
}

func sanitizedAuthenticationError(operation string, err error) error {
	return errors.Errorf("%s failed with code %s", operation, connect.CodeOf(err))
}
