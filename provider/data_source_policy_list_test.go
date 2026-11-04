package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

type policyListTestClient struct {
	api.Client
	response *v1pb.ListPoliciesResponse
	parent   string
}

func (*policyListTestClient) GetWorkspaceName() string {
	return "workspaces/test-workspace"
}

func (c *policyListTestClient) ListPolicies(_ context.Context, parent string) (*v1pb.ListPoliciesResponse, error) {
	c.parent = parent
	return c.response, nil
}

func TestPolicyListIncludesSupportedTypes(t *testing.T) {
	policies := []*v1pb.Policy{
		{Type: v1pb.PolicyType_MASKING_EXEMPTION, Policy: &v1pb.Policy_MaskingExemptionPolicy{MaskingExemptionPolicy: &v1pb.MaskingExemptionPolicy{}}},
		{Type: v1pb.PolicyType_MASKING_RULE, Policy: &v1pb.Policy_MaskingRulePolicy{MaskingRulePolicy: &v1pb.MaskingRulePolicy{}}},
		{Type: v1pb.PolicyType_ROLLOUT_POLICY, Policy: &v1pb.Policy_RolloutPolicy{RolloutPolicy: &v1pb.RolloutPolicy{Automatic: true}}},
		{Type: v1pb.PolicyType_DATA_QUERY, Policy: &v1pb.Policy_QueryDataPolicy{QueryDataPolicy: &v1pb.QueryDataPolicy{MaximumResultRows: 100}}},
		{Type: v1pb.PolicyType_REVIEW_RULE, Policy: &v1pb.Policy_ReviewRulePolicy{ReviewRulePolicy: &v1pb.ReviewRulePolicy{Rules: []v1pb.ReviewRuleType{v1pb.ReviewRuleType_SYNTAX}}}},
	}
	keys := []string{"masking_exemption_policy", "global_masking_policy", "rollout_policy", "query_data_policy", "review_rule_policy"}
	for i, policy := range policies {
		t.Run(policy.Type.String(), func(t *testing.T) {
			policy.Name = fmt.Sprintf("projects/sample/policies/%s", policy.Type)
			c := &policyListTestClient{response: &v1pb.ListPoliciesResponse{Policies: []*v1pb.Policy{policy}}}
			d := schema.TestResourceDataRaw(t, dataSourcePolicyList().Schema, map[string]interface{}{"parent": "projects/sample"})
			if diags := dataSourcePolicyListRead(context.Background(), d, c); diags.HasError() {
				t.Fatal(diags)
			}
			if c.parent != "projects/sample" {
				t.Fatalf("requested parent = %q", c.parent)
			}
			got := d.Get("policies").([]interface{})
			if len(got) != 1 {
				t.Fatalf("policy count = %d, want 1", len(got))
			}
			raw := got[0].(map[string]interface{})
			if raw["type"] != policy.Type.String() || raw["name"] != policy.Name {
				t.Fatalf("policy identity = %v", raw)
			}
			if payload, ok := raw[keys[i]].([]interface{}); !ok || len(payload) != 1 {
				t.Fatalf("policy payload %s = %v", keys[i], raw[keys[i]])
			}
		})
	}
}

func TestPolicyListSkipsUnsupportedTypes(t *testing.T) {
	c := &policyListTestClient{response: &v1pb.ListPoliciesResponse{Policies: []*v1pb.Policy{
		{Type: v1pb.PolicyType_TAG},
		{Type: v1pb.PolicyType(1000)},
		{
			Name:   "workspaces/test-workspace/policies/review_ai",
			Type:   v1pb.PolicyType_REVIEW_AI,
			Policy: &v1pb.Policy_ReviewAiPolicy{ReviewAiPolicy: &v1pb.ReviewAIPolicy{Content: "Require reversible migrations."}},
		},
	}}}
	d := schema.TestResourceDataRaw(t, dataSourcePolicyList().Schema, nil)
	if diags := dataSourcePolicyListRead(context.Background(), d, c); diags.HasError() {
		t.Fatal(diags)
	}
	if c.parent != "workspaces/test-workspace" {
		t.Fatalf("default parent = %q", c.parent)
	}
	if got := d.Get("policies").([]interface{}); len(got) != 0 {
		t.Fatalf("unsupported policies returned: %v", got)
	}
}

func TestAccPolicyListDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			internal.GetTestStepForDataSourceListWithParent(
				"",
				"",
				"bytebase_policy_list",
				"before",
				"policies",
				"projects/project-sample",
				0,
			),
			internal.GetTestStepForDataSourceListWithParent(
				testAccCheckPolicyResource(
					"masking_exemption_policy",
					"projects/project-sample",
					getMaskingExemptionPolicy("instances/test-sample-instance/databases/employee", "salary", "amount"),
					v1pb.PolicyType_MASKING_EXEMPTION,
				),
				"bytebase_policy.masking_exemption_policy",
				"bytebase_policy_list",
				"after",
				"policies",
				"projects/project-sample",
				1,
			),
		},
	})
}
