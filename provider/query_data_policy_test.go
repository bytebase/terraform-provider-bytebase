package provider

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func TestQueryDataPolicyLegacyExportValidation(t *testing.T) {
	for _, value := range []bool{false, true} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			config := queryDataPolicyConfig(100, nil)
			config["query_data_policy"].([]interface{})[0].(map[string]interface{})["disable_export"] = value
			diags := resourcePolicy().Validate(terraform.NewResourceConfigRaw(config))
			if diags.HasError() != value {
				t.Fatalf("disable_export=%v diagnostics = %v", value, diags)
			}
			if value && !strings.Contains(fmt.Sprint(diags), "bb.sql.export") {
				t.Fatalf("missing export permission migration guidance: %v", diags)
			}
		})
	}
}

func TestQueryDataPolicyDownloadLifecycle(t *testing.T) {
	for _, value := range []bool{false, true} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			client := &queryDataPolicyClient{}
			state := applyQueryDataPolicyConfig(t, queryDataPolicyConfig(100, &value), nil, client)
			if client.patch.GetQueryDataPolicy().GetDisableDownload() != value {
				t.Fatalf("created download restriction = %v, want %v", client.patch, value)
			}
			if client.patch.GetQueryDataPolicy().GetDisableExport() { //nolint:staticcheck // The deprecated wire field must stay unset.
				t.Fatal("deprecated disable_export was sent")
			}
			if !slices.Contains(client.masks, "query_data_policy") {
				t.Fatalf("missing query policy mask: %v", client.masks)
			}
			if got := state.Attributes["query_data_policy.0.disable_download"]; got != fmt.Sprint(value) {
				t.Fatalf("created download state = %q", got)
			}
			updated := !value
			state = applyQueryDataPolicyConfig(t, queryDataPolicyConfig(200, &updated), state, client)
			if client.policy.GetQueryDataPolicy().GetDisableDownload() != updated || state.Attributes["query_data_policy.0.disable_download"] != fmt.Sprint(updated) {
				t.Fatalf("updated download restriction: server=%v, state=%v", client.policy, state.Attributes)
			}
			if diff := reviewPolicyDiff(t, resourcePolicy(), queryDataPolicyConfig(200, &updated), state); diff != nil && !diff.Empty() {
				t.Fatalf("unchanged configuration has a diff: %v", diff)
			}
		})
	}
}

func TestQueryDataPolicyOmittedDownloadPreservesServerValue(t *testing.T) {
	client := &queryDataPolicyClient{policy: queryDataPolicyMessage(true)}
	config := queryDataPolicyConfig(100, nil)
	state := applyQueryDataPolicyConfig(t, config, nil, client)
	if !client.policy.GetQueryDataPolicy().GetDisableDownload() {
		t.Fatal("adopting an existing policy cleared its download restriction")
	}
	// Simulate legacy or stale state; omission must not replace a remote true.
	state.Attributes["query_data_policy.0.disable_download"] = "false"
	state = applyQueryDataPolicyConfig(t, queryDataPolicyConfig(200, nil), state, client)
	if !client.patch.GetQueryDataPolicy().GetDisableDownload() || state.Attributes["query_data_policy.0.disable_download"] != "true" {
		t.Fatalf("query policy update cleared an omitted download restriction: patch=%v, state=%v", client.patch, state.Attributes)
	}
}

func TestQueryDataPolicyOmittedDownloadOnNewPolicy(t *testing.T) {
	client := &queryDataPolicyClient{}
	state := applyQueryDataPolicyConfig(t, queryDataPolicyConfig(100, nil), nil, client)
	if client.policy.GetQueryDataPolicy().GetDisableDownload() || state.Attributes["query_data_policy.0.disable_download"] != "false" {
		t.Fatalf("new policy download restriction: server=%v, state=%v", client.policy, state.Attributes)
	}
}

func TestQueryDataPolicyLegacyStateRefresh(t *testing.T) {
	client := &queryDataPolicyClient{policy: queryDataPolicyMessage(true)}
	legacy := &terraform.InstanceState{
		ID: client.policy.Name,
		Attributes: map[string]string{
			"parent":              "workspaces/query-workspace",
			"type":                "DATA_QUERY",
			"name":                client.policy.Name,
			"enforce":             "true",
			"inherit_from_parent": "false",
			"query_data_policy.#": "1",
			"query_data_policy.0.maximum_result_rows":     "100",
			"query_data_policy.0.disable_export":          "true",
			"query_data_policy.0.disable_copy_data":       "false",
			"query_data_policy.0.allow_admin_data_source": "false",
		},
	}
	r := resourcePolicy()
	state, diags := r.RefreshWithoutUpgrade(context.Background(), legacy, client)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if state.Attributes["query_data_policy.0.disable_download"] != "true" || state.Attributes["query_data_policy.0.disable_export"] != "false" {
		t.Fatalf("legacy state was not refreshed to the migrated download restriction: %v", state.Attributes)
	}
	if diff := reviewPolicyDiff(t, r, queryDataPolicyConfig(100, nil), state); diff != nil && !diff.Empty() {
		t.Fatalf("removing the retired flag requires no state migration or policy change: %v", diff)
	}
	state = applyQueryDataPolicyConfig(t, queryDataPolicyConfig(200, nil), state, client)
	if !client.policy.GetQueryDataPolicy().GetDisableDownload() || state.Attributes["query_data_policy.0.disable_download"] != "true" {
		t.Fatalf("updating the legacy policy lost the download restriction: %v", state.Attributes)
	}
}

func TestQueryDataPolicyDownloadReadImportAndList(t *testing.T) {
	client := &queryDataPolicyClient{policy: queryDataPolicyMessage(true)}
	for name, r := range map[string]*schema.Resource{"resource": resourcePolicy(), "data_source": dataSourcePolicy()} {
		t.Run(name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"parent": "workspaces/query-workspace", "type": "DATA_QUERY"})
			d.SetId(client.policy.Name)
			if r.Importer != nil {
				if _, err := r.Importer.StateContext(context.Background(), d, client); err != nil {
					t.Fatal(err)
				}
			}
			if diags := r.ReadContext(context.Background(), d, client); diags.HasError() {
				t.Fatal(diags)
			}
			if got, ok := d.Get("query_data_policy.0.disable_download").(bool); !ok || !got {
				t.Fatalf("read download restriction = %v, want true", got)
			}
		})
	}
	r := dataSourcePolicyList()
	d := schema.TestResourceDataRaw(t, r.Schema, nil)
	if diags := r.ReadContext(context.Background(), d, client); diags.HasError() {
		t.Fatal(diags)
	}
	policies := d.Get("policies").([]interface{})
	block := policies[0].(map[string]interface{})["query_data_policy"].([]interface{})[0].(map[string]interface{})
	if download, ok := block["disable_download"].(bool); !ok || !download {
		t.Fatalf("listed download restriction = %v", block)
	}
}

func TestQueryDataPolicyPreservationReadFailureStopsWrite(t *testing.T) {
	for _, message := range []string{"cannot read existing policy", "not_found is not visible to this caller"} {
		t.Run(message, func(t *testing.T) {
			client := &queryDataPolicyClient{readError: connect.NewError(connect.CodePermissionDenied, errors.New(message))}
			r := resourcePolicy()
			diff := reviewPolicyDiff(t, r, queryDataPolicyConfig(100, nil), nil)
			if _, diags := r.Apply(context.Background(), nil, diff, client); !diags.HasError() {
				t.Fatal("policy write succeeded without reading the omitted download restriction")
			}
			if client.patch != nil {
				t.Fatalf("policy was written despite preservation read failure: %v", client.patch)
			}
		})
	}
}

func queryDataPolicyConfig(rows int, download *bool) map[string]interface{} {
	block := map[string]interface{}{"maximum_result_rows": rows}
	if download != nil {
		block["disable_download"] = *download
	}
	return map[string]interface{}{
		"parent":            "workspaces/query-workspace",
		"type":              "DATA_QUERY",
		"query_data_policy": []interface{}{block},
	}
}

func applyQueryDataPolicyConfig(t *testing.T, config map[string]interface{}, state *terraform.InstanceState, client *queryDataPolicyClient) *terraform.InstanceState {
	t.Helper()
	r := resourcePolicy()
	diff := reviewPolicyDiff(t, r, config, state)
	result, diags := r.Apply(context.Background(), state, diff, client)
	if diags.HasError() {
		t.Fatalf("query policy apply failed: %v", diags)
	}
	return result
}

func queryDataPolicyMessage(download bool) *v1pb.Policy {
	return &v1pb.Policy{
		Name:    "workspaces/query-workspace/policies/data_query",
		Type:    v1pb.PolicyType_DATA_QUERY,
		Enforce: true,
		Policy: &v1pb.Policy_QueryDataPolicy{QueryDataPolicy: &v1pb.QueryDataPolicy{
			MaximumResultRows: 100,
			DisableDownload:   download,
		}},
	}
}

type queryDataPolicyClient struct {
	api.Client
	policy    *v1pb.Policy
	patch     *v1pb.Policy
	masks     []string
	readError error
}

func (*queryDataPolicyClient) GetWorkspaceName() string { return "workspaces/query-workspace" }

func (c *queryDataPolicyClient) GetPolicy(context.Context, string) (*v1pb.Policy, error) {
	if c.readError != nil {
		return nil, c.readError
	}
	if c.policy == nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("policy not found"))
	}
	return proto.Clone(c.policy).(*v1pb.Policy), nil
}

func (c *queryDataPolicyClient) UpsertPolicy(_ context.Context, patch *v1pb.Policy, masks []string) (*v1pb.Policy, error) {
	c.patch = proto.Clone(patch).(*v1pb.Policy)
	c.masks = slices.Clone(masks)
	if c.policy == nil {
		c.policy = queryDataPolicyMessage(false)
	}
	if slices.Contains(masks, "query_data_policy") {
		c.policy.Policy = proto.Clone(patch).(*v1pb.Policy).Policy
	}
	if slices.Contains(masks, "enforce") {
		c.policy.Enforce = patch.Enforce
	}
	return proto.Clone(c.policy).(*v1pb.Policy), nil
}

func (c *queryDataPolicyClient) ListPolicies(context.Context, string) (*v1pb.ListPoliciesResponse, error) {
	return &v1pb.ListPoliciesResponse{Policies: []*v1pb.Policy{proto.Clone(c.policy).(*v1pb.Policy)}}, nil
}
