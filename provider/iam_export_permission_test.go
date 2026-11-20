package provider

import (
	"context"
	"slices"
	"strings"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func TestConvertToIAMPolicyWithScopedRole(t *testing.T) {
	const expiringScope = `resource.database == "instances/reporting/databases/sales" && resource.schema_name == "public" && resource.table_name in ["orders"] && request.time < timestamp("2035-01-02T03:04:05Z")`
	for _, tc := range []struct {
		name, role, permission string
		environmentIDs         []interface{}
		wantCondition          string
		wantErr                bool
	}{
		{
			name: "export-only role", role: "roles/reportExporter", permission: "bb.sql.export",
			wantCondition: expiringScope + ` && resource.environment_id in []`,
		},
		{
			name: "select-only role", role: "roles/reportReader", permission: "bb.sql.select",
			wantCondition: expiringScope + ` && resource.environment_id in []`,
		},
		{
			name: "DDL role with environment", role: "roles/reportEditor", permission: "bb.sql.ddl",
			environmentIDs: []interface{}{"test"},
			wantCondition:  expiringScope + ` && resource.environment_id in ["test"]`,
		},
		{name: "non-SQL role", role: "roles/databaseViewer", permission: "bb.databases.list", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const member = "user:analyst@example.com"
			condition := map[string]interface{}{
				"database":         "instances/reporting/databases/sales",
				"schema":           "public",
				"tables":           []interface{}{"orders"},
				"expire_timestamp": "2035-01-02T03:04:05Z",
			}
			if tc.environmentIDs != nil {
				condition["environment_ids"] = tc.environmentIDs
			}
			data := schema.TestResourceDataRaw(t, resourceIAMPolicy().Schema, map[string]interface{}{
				"parent": "projects/analytics",
				"iam_policy": []interface{}{map[string]interface{}{
					"binding": []interface{}{map[string]interface{}{
						"role":      tc.role,
						"members":   []interface{}{member},
						"condition": []interface{}{condition},
					}},
				}},
			})
			client := &iamPolicyRoleClient{role: &v1pb.Role{
				Name:        tc.role,
				Permissions: []string{tc.permission},
			}}

			policy, err := convertToIAMPolicy(t.Context(), client, data)
			if tc.wantErr {
				if err == nil || !strings.Contains(err.Error(), "shouldn't configure the database condition") {
					t.Fatalf("expected database condition rejection for %s, got %v", tc.role, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(policy.Bindings) != 1 {
				t.Fatalf("got %d bindings, want 1", len(policy.Bindings))
			}
			binding := policy.Bindings[0]
			if binding.Role != tc.role || !slices.Equal(binding.Members, []string{member}) {
				t.Errorf("got role %q with members %v, want %q with member %q", binding.Role, binding.Members, tc.role, member)
			}
			if got := binding.GetCondition().GetExpression(); got != tc.wantCondition {
				t.Errorf("got condition %q, want %q", got, tc.wantCondition)
			}
		})
	}
}

type iamPolicyRoleClient struct {
	api.Client
	role *v1pb.Role
}

func (c *iamPolicyRoleClient) GetRole(context.Context, string) (*v1pb.Role, error) {
	return c.role, nil
}
