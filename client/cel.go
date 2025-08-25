package client

import (
	"context"
	"fmt"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/pkg/errors"
	v1alpha1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// ParseExpression parse the expression string using Connect RPC.
func (c *client) ParseExpression(ctx context.Context, expression string) (*v1alpha1.Expr, error) {
	if c.celClient == nil {
		return nil, fmt.Errorf("cel service client not initialized")
	}

	req := connect.NewRequest(&v1pb.BatchParseRequest{
		Expressions: []string{expression},
	})

	resp, err := c.celClient.BatchParse(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Msg.Expressions) != 1 {
		return nil, errors.Errorf("failed to parse the cel: %v", expression)
	}

	return resp.Msg.GetExpressions()[0], nil
}
