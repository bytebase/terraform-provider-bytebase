package provider

import (
	"context"
	"slices"
	"strings"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func TestReviewAIPolicyUnsupported(t *testing.T) {
	for name, r := range map[string]*schema.Resource{"resource": resourcePolicy(), "data_source": dataSourcePolicy()} {
		t.Run(name, func(t *testing.T) {
			config := terraform.NewResourceConfigRaw(map[string]interface{}{"type": "REVIEW_AI"})
			if diags := r.Validate(config); !diags.HasError() {
				t.Fatal("dark-launched REVIEW_AI policy type was accepted")
			}
		})
	}
}

func TestReviewAIPolicySchemasAbsent(t *testing.T) {
	for name, r := range map[string]*schema.Resource{
		"resource":    resourcePolicy(),
		"data_source": dataSourcePolicy(),
		"list":        dataSourcePolicyList().Schema["policies"].Elem.(*schema.Resource),
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := r.Schema["review_ai_policy"]; ok {
				t.Fatal("dark-launched review_ai_policy is exposed in the schema")
			}
		})
	}
}

func TestReviewAIPolicyReadUnsupported(t *testing.T) {
	r := resourcePolicy()
	data := schema.TestResourceDataRaw(t, r.Schema, nil)
	data.SetId("projects/sample/policies/review_ai")
	client := &reviewPolicyClient{policy: &v1pb.Policy{
		Name: "projects/sample/policies/review_ai",
		Type: v1pb.PolicyType_REVIEW_AI,
		Policy: &v1pb.Policy_ReviewAiPolicy{ReviewAiPolicy: &v1pb.ReviewAIPolicy{
			Content: "Require a rollback plan.",
		}},
	}}
	if _, err := r.Importer.StateContext(context.Background(), data, client); err != nil {
		t.Fatal(err)
	}
	if diags := r.ReadContext(context.Background(), data, client); !diags.HasError() {
		t.Fatal("imported dark-launched REVIEW_AI policy was read into state")
	}
}

func TestReviewPolicySchemas(t *testing.T) {
	for name, r := range map[string]*schema.Resource{"resource": resourcePolicy(), "data_source": dataSourcePolicy()} {
		for _, policyType := range []v1pb.PolicyType{v1pb.PolicyType_REVIEW_RULE} {
			t.Run(name+"/"+policyType.String(), func(t *testing.T) {
				field := strings.ToLower(policyType.String()) + "_policy"
				if r.Schema[field] == nil {
					t.Fatalf("missing %s", field)
				}
				if _, errs := r.Schema["type"].ValidateFunc(policyType.String(), "type"); len(errs) != 0 {
					t.Fatalf("type rejects %s: %v", policyType, errs)
				}
			})
		}
	}
}

func TestReviewPolicyLifecycle(t *testing.T) {
	for _, policyType := range []v1pb.PolicyType{v1pb.PolicyType_REVIEW_RULE} {
		for _, parent := range []string{"", "projects/review-project"} {
			t.Run(policyType.String()+"/"+parent, func(t *testing.T) {
				r := resourcePolicy()
				field := strings.ToLower(policyType.String()) + "_policy"
				if r.Schema[field] == nil {
					t.Fatalf("missing %s", field)
				}
				client := &reviewPolicyClient{}
				config := reviewPolicyConfig(policyType, parent, false)
				state := applyReviewPolicyConfig(t, r, config, nil, client)
				resolvedParent := parent
				if resolvedParent == "" {
					resolvedParent = client.GetWorkspaceName()
				}
				if got := client.patch.Name; got != resolvedParent+"/policies/"+policyType.String() {
					t.Fatalf("create name = %q", got)
				}
				if client.patch.Type != policyType || !slices.Contains(client.masks, field) {
					t.Fatalf("create request = %v, masks = %v", client.patch, client.masks)
				}
				if !client.policy.Enforce || client.policy.InheritFromParent {
					t.Fatalf("unexpected policy defaults: %v", client.policy)
				}
				assertReviewPolicyPayload(t, policyType, client.patch, false)
				assertReviewPolicyState(t, state, field, false)
				if diff := reviewPolicyDiff(t, r, config, state); diff != nil && !diff.Empty() {
					t.Fatalf("unchanged config has a diff: %v", diff)
				}

				config = reviewPolicyConfig(policyType, parent, true)
				state = applyReviewPolicyConfig(t, r, config, state, client)
				if !slices.Contains(client.masks, field) {
					t.Fatalf("payload update missing %s in %v", field, client.masks)
				}
				assertReviewPolicyPayload(t, policyType, client.patch, true)
				assertReviewPolicyState(t, state, field, true)

				config["enforce"] = false
				state = applyReviewPolicyConfig(t, r, config, state, client)
				if !slices.Equal(client.masks, []string{"enforce"}) || client.policy.Enforce {
					t.Fatalf("enforce-only update: masks = %v, policy = %v", client.masks, client.policy)
				}
				assertReviewPolicyPayload(t, policyType, client.policy, true)

				ds := dataSourcePolicy()
				data := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{"parent": parent, "type": policyType.String()})
				if diags := ds.ReadContext(context.Background(), data, client); diags.HasError() {
					t.Fatalf("data source read: %v", diags)
				}
				assertReviewPolicyState(t, data.State(), field, true)

				imported := schema.TestResourceDataRaw(t, r.Schema, nil)
				imported.SetId(state.ID)
				if _, err := r.Importer.StateContext(context.Background(), imported, client); err != nil {
					t.Fatalf("import: %v", err)
				}
				if diags := r.ReadContext(context.Background(), imported, client); diags.HasError() {
					t.Fatalf("import read: %v", diags)
				}
				if imported.Get("parent") != resolvedParent || imported.Get("type") != policyType.String() {
					t.Fatalf("import parent/type = %v/%v", imported.Get("parent"), imported.Get("type"))
				}
				assertReviewPolicyState(t, imported.State(), field, true)

				deleted, diags := r.Apply(context.Background(), state, &terraform.InstanceDiff{Destroy: true}, client)
				if diags.HasError() || deleted != nil && deleted.ID != "" || client.deleted != state.ID {
					t.Fatalf("delete: state = %v, diagnostics = %v, deleted = %q", deleted, diags, client.deleted)
				}
				if diags := r.ReadContext(context.Background(), imported, client); diags.HasError() || imported.Id() != "" {
					t.Fatalf("missing policy was not removed from state: id = %q, diagnostics = %v", imported.Id(), diags)
				}
			})
		}
	}
}

func TestReviewPolicyValidation(t *testing.T) {
	for _, tc := range []struct {
		name       string
		policyType v1pb.PolicyType
		parent     string
		payload    map[string]interface{}
	}{
		{"unknown rule", v1pb.PolicyType_REVIEW_RULE, "", map[string]interface{}{"rules": []interface{}{"UNKNOWN"}}},
		{"unspecified rule", v1pb.PolicyType_REVIEW_RULE, "", map[string]interface{}{"rules": []interface{}{"REVIEW_RULE_TYPE_UNSPECIFIED"}}},
		{"duplicate rule", v1pb.PolicyType_REVIEW_RULE, "", map[string]interface{}{"rules": []interface{}{"SYNTAX", "SYNTAX"}}},
		{"rule environment", v1pb.PolicyType_REVIEW_RULE, "environments/prod", map[string]interface{}{"rules": []interface{}{"SYNTAX"}}},
		{"rule instance", v1pb.PolicyType_REVIEW_RULE, "instances/mysql", map[string]interface{}{"rules": []interface{}{"SYNTAX"}}},
		{"rule database", v1pb.PolicyType_REVIEW_RULE, "instances/mysql/databases/app", map[string]interface{}{"rules": []interface{}{"SYNTAX"}}},
		{"missing rule block", v1pb.PolicyType_REVIEW_RULE, "", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := resourcePolicy()
			field := strings.ToLower(tc.policyType.String()) + "_policy"
			if r.Schema[field] == nil {
				t.Fatalf("missing %s", field)
			}
			config := map[string]interface{}{"type": tc.policyType.String(), "parent": tc.parent}
			if tc.payload != nil {
				config[field] = []interface{}{tc.payload}
			}
			terraformConfig := terraform.NewResourceConfigRaw(config)
			if diags := r.Validate(terraformConfig); diags.HasError() {
				return
			}
			if _, err := r.Diff(context.Background(), nil, terraformConfig, nil); err == nil {
				t.Fatal("invalid review policy was accepted during planning")
			}
		})
	}
}

func TestReviewPolicyPayloadType(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config map[string]interface{}
	}{
		{"query with rule block", map[string]interface{}{
			"type":               "DATA_QUERY",
			"query_data_policy":  []interface{}{map[string]interface{}{}},
			"review_rule_policy": []interface{}{map[string]interface{}{"rules": []interface{}{"SYNTAX"}}},
		}},
		{"rules with masking block", map[string]interface{}{
			"type":                  "REVIEW_RULE",
			"review_rule_policy":    []interface{}{map[string]interface{}{"rules": []interface{}{"SYNTAX"}}},
			"global_masking_policy": []interface{}{map[string]interface{}{}},
		}},
		{"rules with exemption block", map[string]interface{}{
			"type":                     "REVIEW_RULE",
			"review_rule_policy":       []interface{}{map[string]interface{}{"rules": []interface{}{"SYNTAX"}}},
			"masking_exemption_policy": []interface{}{map[string]interface{}{}},
		}},
		{"rules with rollout block", map[string]interface{}{
			"type":               "REVIEW_RULE",
			"review_rule_policy": []interface{}{map[string]interface{}{"rules": []interface{}{"SYNTAX"}}},
			"rollout_policy":     []interface{}{map[string]interface{}{}},
		}},
		{"rules with query block", map[string]interface{}{
			"type":               "REVIEW_RULE",
			"review_rule_policy": []interface{}{map[string]interface{}{"rules": []interface{}{"SYNTAX"}}},
			"query_data_policy":  []interface{}{map[string]interface{}{}},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := resourcePolicy().Diff(context.Background(), nil, terraform.NewResourceConfigRaw(tc.config), nil); err == nil {
				t.Fatal("conflicting policy blocks were accepted")
			}
		})
	}
}

func TestReviewPolicyUnknownType(t *testing.T) {
	r := resourcePolicy()
	config := reviewPolicyConfig(v1pb.PolicyType_REVIEW_RULE, "", false)
	value, err := schema.JSONMapToStateValue(config, r.CoreConfigSchema())
	if err != nil {
		t.Fatal(err)
	}
	attrs := value.AsValueMap()
	attrs["type"] = cty.UnknownVal(cty.String)
	raw := terraform.NewResourceConfigShimmed(cty.ObjectVal(attrs), r.CoreConfigSchema())
	if _, err := r.Diff(context.Background(), nil, raw, nil); err != nil {
		t.Fatalf("unknown policy type must be deferred until apply: %v", err)
	}
}

func TestReviewPolicyUnknownPayload(t *testing.T) {
	for _, tc := range []struct {
		name       string
		policyType v1pb.PolicyType
		block      cty.Value
	}{
		{"rules list", v1pb.PolicyType_REVIEW_RULE, cty.ObjectVal(map[string]cty.Value{"rules": cty.UnknownVal(cty.List(cty.String))})},
		{"rule entry", v1pb.PolicyType_REVIEW_RULE, cty.ObjectVal(map[string]cty.Value{"rules": cty.ListVal([]cty.Value{cty.StringVal("SYNTAX"), cty.UnknownVal(cty.String)})})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := resourcePolicy()
			value, err := schema.JSONMapToStateValue(reviewPolicyConfig(tc.policyType, "", false), r.CoreConfigSchema())
			if err != nil {
				t.Fatal(err)
			}
			attrs := value.AsValueMap()
			field := strings.ToLower(tc.policyType.String()) + "_policy"
			attrs[field] = cty.ListVal([]cty.Value{tc.block})
			value = cty.ObjectVal(attrs)
			raw := terraform.NewResourceConfigShimmed(value, r.CoreConfigSchema())
			if _, err := r.Diff(context.Background(), &terraform.InstanceState{RawConfig: value}, raw, nil); err != nil {
				t.Fatalf("unknown %s must be deferred until apply: %v", tc.name, err)
			}
		})
	}
}

func reviewPolicyConfig(policyType v1pb.PolicyType, parent string, updated bool) map[string]interface{} {
	config := map[string]interface{}{"type": policyType.String()}
	if parent != "" {
		config["parent"] = parent
	}
	if policyType == v1pb.PolicyType_REVIEW_RULE {
		rules := []interface{}{"SYNTAX", "REQUIRE_WHERE"}
		if updated {
			rules = []interface{}{}
		}
		config["review_rule_policy"] = []interface{}{map[string]interface{}{"rules": rules}}
	}
	return config
}

func assertReviewPolicyPayload(t *testing.T, policyType v1pb.PolicyType, policy *v1pb.Policy, updated bool) {
	t.Helper()
	if policyType == v1pb.PolicyType_REVIEW_RULE {
		want := []v1pb.ReviewRuleType{v1pb.ReviewRuleType_SYNTAX, v1pb.ReviewRuleType_REQUIRE_WHERE}
		if updated {
			want = []v1pb.ReviewRuleType{}
		}
		if policy.GetReviewRulePolicy() == nil || !slices.Equal(policy.GetReviewRulePolicy().Rules, want) {
			t.Fatalf("review rule payload = %v, want %v", policy.GetReviewRulePolicy(), want)
		}
	}
}

func assertReviewPolicyState(t *testing.T, state *terraform.InstanceState, field string, updated bool) {
	t.Helper()
	if state.Attributes[field+".#"] != "1" {
		t.Fatalf("state missing %s: %v", field, state.Attributes)
	}
	if field == "review_rule_policy" {
		want := "2"
		if updated {
			want = "0"
		}
		if got := state.Attributes[field+".0.rules.#"]; got != want {
			t.Fatalf("rules count = %q, want %q", got, want)
		}
	}
}

func reviewPolicyDiff(t *testing.T, r *schema.Resource, config map[string]interface{}, state *terraform.InstanceState) *terraform.InstanceDiff {
	t.Helper()
	value, err := schema.JSONMapToStateValue(config, r.CoreConfigSchema())
	if err != nil {
		t.Fatalf("coerce policy config: %v", err)
	}
	raw := terraform.NewResourceConfigRaw(config)
	raw.CtyValue = value
	diff, err := r.Diff(context.Background(), state, raw, nil)
	if err != nil {
		t.Fatalf("policy diff: %v", err)
	}
	if diff != nil {
		diff.RawConfig = value
	}
	return diff
}

func applyReviewPolicyConfig(t *testing.T, r *schema.Resource, config map[string]interface{}, state *terraform.InstanceState, client *reviewPolicyClient) *terraform.InstanceState {
	t.Helper()
	diff := reviewPolicyDiff(t, r, config, state)
	updated, diags := r.Apply(context.Background(), state, diff, client)
	if diags.HasError() {
		t.Fatalf("apply policy: %v", diags)
	}
	return updated
}

type reviewPolicyClient struct {
	api.Client
	policy  *v1pb.Policy
	patch   *v1pb.Policy
	masks   []string
	deleted string
}

func (*reviewPolicyClient) GetWorkspaceName() string { return "workspaces/review-workspace" }

func (c *reviewPolicyClient) GetPolicy(_ context.Context, name string) (*v1pb.Policy, error) {
	if c.policy == nil || !strings.EqualFold(c.policy.Name, name) {
		return nil, connect.NewError(connect.CodeNotFound, errors.Errorf("policy %s not found", name))
	}
	return proto.Clone(c.policy).(*v1pb.Policy), nil
}

func (c *reviewPolicyClient) UpsertPolicy(_ context.Context, patch *v1pb.Policy, masks []string) (*v1pb.Policy, error) {
	c.patch = proto.Clone(patch).(*v1pb.Policy)
	c.masks = slices.Clone(masks)
	if c.policy == nil {
		c.policy = proto.Clone(patch).(*v1pb.Policy)
		c.policy.Enforce = true
		parent, _, _ := strings.Cut(patch.Name, "/policies/")
		c.policy.Name = parent + "/policies/" + strings.ToLower(patch.Type.String())
	} else {
		for _, mask := range masks {
			switch mask {
			case "review_rule_policy":
				c.policy.Policy = proto.Clone(patch).(*v1pb.Policy).Policy
			case "enforce":
				c.policy.Enforce = patch.Enforce
			case "inherit_from_parent":
				c.policy.InheritFromParent = patch.InheritFromParent
			default:
			}
		}
	}
	return proto.Clone(c.policy).(*v1pb.Policy), nil
}

func (c *reviewPolicyClient) DeletePolicy(_ context.Context, name string) error {
	c.deleted = name
	c.policy = nil
	return nil
}
