package client

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"connectrpc.com/connect"
	"github.com/pkg/errors"
)

// Note: The login method has been moved to client.go and now uses Connect RPC.
// This file is kept for backward compatibility but the implementation
// has been migrated to use the AuthServiceClient from Connect RPC.
// authInterceptor implements connect.Interceptor to add authentication headers.
type authInterceptor struct {
	tokenManager  *tokenManager
	customHeaders map[string]string
}

type tokenManager struct {
	mu            sync.RWMutex
	token         string
	authenticator authenticator
}

func (m *tokenManager) Token() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.token
}

func (m *tokenManager) RefreshIfCurrent(ctx context.Context, failedToken string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.token != failedToken {
		return m.token, nil
	}
	token, err := m.authenticator.Authenticate(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to refresh authentication")
	}
	m.token = token
	return token, nil
}

func setHeaders(dst http.Header, headers map[string]string) {
	for name, value := range headers {
		dst.Set(name, value)
	}
}

func (a *authInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if !req.Spec().IsClient {
			return next(ctx, req)
		}
		setHeaders(req.Header(), a.customHeaders)
		failedToken := a.tokenManager.Token()
		if failedToken != "" {
			req.Header().Set("Authorization", fmt.Sprintf("Bearer %s", failedToken))
		}
		resp, err := next(ctx, req)
		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			return resp, err
		}

		refreshedToken, refreshErr := a.tokenManager.RefreshIfCurrent(ctx, failedToken)
		if refreshErr != nil {
			return nil, refreshErr
		}
		req.Header().Set("Authorization", fmt.Sprintf("Bearer %s", refreshedToken))
		return next(ctx, req)
	})
}

func (a *authInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		setHeaders(conn.RequestHeader(), a.customHeaders)
		if token := a.tokenManager.Token(); token != "" {
			conn.RequestHeader().Set("Authorization", fmt.Sprintf("Bearer %s", token))
		}
		return conn
	})
}

func (*authInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	})
}
