package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	"github.com/pkg/errors"
	v1alpha1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/encoding/protojson"
)

// ParseExpression parse the expression string.
func (c *client) ParseExpression(ctx context.Context, expression string) (*v1alpha1.Expr, error) {
	payload, err := protojson.Marshal(&v1pb.BatchParseRequest{
		Expressions: []string{expression},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/cel/batchParse", c.url, c.version), strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, fmt.Sprintf("parse cel response:\n%v", string(body)))

	var res v1pb.BatchParseResponse
	if err := ProtojsonUnmarshaler.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	if len(res.Expressions) != 1 {
		return nil, errors.Errorf("failed to parse the cel: %v", expression)
	}

	return res.GetExpressions()[0], nil
}
