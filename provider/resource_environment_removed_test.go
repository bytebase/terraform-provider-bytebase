package provider

import (
	"context"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

type environmentSettingClient struct {
	api.Client
	environments []*v1pb.EnvironmentSetting_Environment
	upserts      int
}

func (c *environmentSettingClient) GetSetting(_ context.Context, name string) (*v1pb.Setting, error) {
	return &v1pb.Setting{
		Name: name,
		Value: &v1pb.SettingValue{
			Value: &v1pb.SettingValue_Environment{
				Environment: &v1pb.EnvironmentSetting{Environments: c.environments},
			},
		},
	}, nil
}

func (c *environmentSettingClient) UpsertSetting(_ context.Context, upsert *v1pb.Setting, _ []string) (*v1pb.Setting, error) {
	c.upserts++
	c.environments = upsert.GetValue().GetEnvironment().GetEnvironments()
	return upsert, nil
}

func newRemovedEnvironmentData(t *testing.T) (*schema.ResourceData, *environmentSettingClient) {
	t.Helper()
	d := schema.TestResourceDataRaw(t, resourceEnvironment().Schema, map[string]interface{}{
		"resource_id": "test",
		"title":       "Test",
	})
	d.SetId("environments/test")
	client := &environmentSettingClient{
		environments: []*v1pb.EnvironmentSetting_Environment{
			{Id: "prod", Name: "environments/prod", Title: "Prod"},
		},
	}
	return d, client
}

func TestResourceEnvironmentReadRemovesMissingEnvironment(t *testing.T) {
	d, client := newRemovedEnvironmentData(t)

	diags := resourceEnvironmentRead(context.Background(), d, client)
	if diags.HasError() {
		t.Fatalf("resourceEnvironmentRead returned error: %v", diags)
	}
	if d.Id() != "" {
		t.Fatalf("resource id = %q, want it removed from state", d.Id())
	}
}

func TestResourceEnvironmentDeleteMissingEnvironment(t *testing.T) {
	d, client := newRemovedEnvironmentData(t)

	diags := resourceEnvironmentDelete(context.Background(), d, client)
	if diags.HasError() {
		t.Fatalf("resourceEnvironmentDelete returned error: %v", diags)
	}
	if d.Id() != "" {
		t.Fatalf("resource id = %q, want it removed from state", d.Id())
	}
	if client.upserts != 0 {
		t.Fatalf("UpsertSetting called %d times, want 0", client.upserts)
	}
	if len(client.environments) != 1 {
		t.Fatalf("environments = %v, want the unrelated environment kept", client.environments)
	}
}
