package client

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
)

// Note: The login method has been moved to client.go and now uses Connect RPC.
// This file is kept for backward compatibility but the implementation
// has been migrated to use the AuthServiceClient from Connect RPC.
// authInterceptor implements connect.Interceptor to add authentication headers.
type authInterceptor struct {
	token         string
	customHeaders map[string]string
}

func setHeaders(dst http.Header, headers map[string]string) {
	for name, value := range headers {
		dst.Set(name, value)
	}
}

func (a *authInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if req.Spec().IsClient {
			setHeaders(req.Header(), a.customHeaders)
			if a.token != "" {
				req.Header().Set("Authorization", fmt.Sprintf("Bearer %s", a.token))
			}
		}
		return next(ctx, req)
	})
}

func (a *authInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		setHeaders(conn.RequestHeader(), a.customHeaders)
		if a.token != "" {
			conn.RequestHeader().Set("Authorization", fmt.Sprintf("Bearer %s", a.token))
		}
		return conn
	})
}

func (*authInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	})
}
