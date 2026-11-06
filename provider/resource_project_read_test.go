package provider

import (
	"context"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

type projectReadClient struct {
	api.Client
	project *v1pb.Project
}

func (c *projectReadClient) GetProject(_ context.Context, name string) (*v1pb.Project, error) {
	if c.project == nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.Errorf("project %q not found", name))
	}
	return c.project, nil
}

func TestResourceProjectReadRemovesMissingProject(t *testing.T) {
	tests := []struct {
		name    string
		project *v1pb.Project
	}{
		{
			name: "purged",
		},
		{
			name: "archived",
			project: &v1pb.Project{
				Name:  "projects/sample",
				Title: "Sample",
				State: v1pb.State_DELETED,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceProjct().Schema, map[string]interface{}{
				"resource_id": "sample",
				"title":       "Sample",
			})
			d.SetId("projects/sample")

			diags := resourceProjectRead(context.Background(), d, &projectReadClient{project: tc.project})
			if diags.HasError() {
				t.Fatalf("resourceProjectRead returned error: %v", diags)
			}
			if d.Id() != "" {
				t.Fatalf("resource id = %q, want it removed from state", d.Id())
			}
		})
	}
}
