package provider

import (
	"context"
	"slices"
	"strings"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func supportedPolicyTypes() []string {
	return []string{
		v1pb.PolicyType_MASKING_EXEMPTION.String(),
		v1pb.PolicyType_MASKING_RULE.String(),
		v1pb.PolicyType_ROLLOUT_POLICY.String(),
		v1pb.PolicyType_DATA_QUERY.String(),
		v1pb.PolicyType_REVIEW_RULE.String(),
		v1pb.PolicyType_REVIEW_AI.String(),
	}
}

func isSupportedPolicyType(policyType v1pb.PolicyType) bool {
	return slices.Contains(supportedPolicyTypes(), policyType.String())
}

func getReviewRulePolicySchema(computed bool) *schema.Schema {
	rules := &schema.Schema{
		Type:        schema.TypeList,
		Computed:    computed,
		Optional:    !computed,
		Description: "The standard review rules switched on. An empty list switches every rule off. Without SYNTAX, every rule is off.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
	if !computed {
		allowedRules := []string{}
		for _, value := range v1pb.ReviewRuleType_name {
			if value != v1pb.ReviewRuleType_REVIEW_RULE_TYPE_UNSPECIFIED.String() {
				allowedRules = append(allowedRules, value)
			}
		}
		slices.Sort(allowedRules)
		rules.Elem.(*schema.Schema).ValidateFunc = validation.StringInSlice(allowedRules, false)
	}
	policy := &schema.Schema{
		Type:        schema.TypeList,
		Computed:    computed,
		Optional:    !computed,
		Description: "The standard review rule policy. Supported for workspaces and projects. A project policy overrides the workspace policy.",
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"rules": rules,
		}},
	}
	if !computed {
		policy.MaxItems = 1
	}
	return policy
}

const maxReviewAIPolicyBytes = 64 << 10

func getReviewAIPolicySchema(computed bool) *schema.Schema {
	content := &schema.Schema{
		Type:        schema.TypeString,
		Required:    !computed,
		Computed:    computed,
		Description: "Nonblank natural-language SQL review instructions, up to 65,536 UTF-8 bytes.",
	}
	if !computed {
		content.ValidateFunc = func(value interface{}, key string) ([]string, []error) {
			if err := validateReviewAIPolicyContent(value.(string)); err != nil {
				return nil, []error{errors.Wrap(err, key)}
			}
			return nil, nil
		}
	}
	return singletonBlock(computed, "Optional custom instructions for AI SQL review. Supported for workspaces and projects. Both policies apply when enforced, and project instructions take precedence on conflicts. Not required to enable AI SQL review.", map[string]*schema.Schema{
		"content": content,
	})
}

func reviewPolicyBlockName(policyType v1pb.PolicyType) string {
	switch policyType {
	case v1pb.PolicyType_REVIEW_RULE:
		return "review_rule_policy"
	case v1pb.PolicyType_REVIEW_AI:
		return "review_ai_policy"
	default:
		return ""
	}
}

func validateReviewPolicyParent(policyType v1pb.PolicyType, parent string) error {
	if reviewPolicyBlockName(policyType) == "" {
		return nil
	}
	if parent == "" || strings.HasPrefix(parent, internal.WorkspaceNamePrefix) || strings.HasPrefix(parent, internal.ProjectNamePrefix) {
		return nil
	}
	return errors.Errorf("policy %s only supports workspace or project resources", policyType)
}

func validateReviewPolicyBlocks(policyType v1pb.PolicyType, hasBlock func(string) bool) error {
	if reviewPolicyBlockName(policyType) != "" {
		for _, field := range []string{"masking_exemption_policy", "global_masking_policy", "rollout_policy", "query_data_policy"} {
			if hasBlock(field) {
				return errors.Errorf("%s is not valid for policy type %s", field, policyType)
			}
		}
	}
	for _, reviewType := range []v1pb.PolicyType{v1pb.PolicyType_REVIEW_RULE, v1pb.PolicyType_REVIEW_AI} {
		field := reviewPolicyBlockName(reviewType)
		if policyType != reviewType && hasBlock(field) {
			return errors.Errorf("%s requires policy type %s", field, reviewType)
		}
	}
	return nil
}

func customizeReviewPolicyDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	if !d.NewValueKnown("type") {
		return nil
	}
	policyType := v1pb.PolicyType(v1pb.PolicyType_value[d.Get("type").(string)])
	if d.NewValueKnown("parent") {
		if err := validateReviewPolicyParent(policyType, d.Get("parent").(string)); err != nil {
			return err
		}
	}
	if err := validateReviewPolicyBlocks(policyType, func(field string) bool {
		return d.NewValueKnown(field) && len(d.Get(field).([]interface{})) > 0
	}); err != nil {
		return err
	}
	field := reviewPolicyBlockName(policyType)
	if field == "" || !d.NewValueKnown(field) {
		return nil
	}
	if rawConfig := d.GetRawConfig(); !rawConfig.IsNull() && !rawConfig.GetAttr(field).IsWhollyKnown() {
		return nil
	}
	switch policyType {
	case v1pb.PolicyType_REVIEW_RULE:
		_, err := expandReviewRulePolicy(d.Get(field))
		return err
	case v1pb.PolicyType_REVIEW_AI:
		_, err := expandReviewAIPolicy(d.Get(field))
		return err
	default:
		return nil
	}
}

func convertToReviewRulePolicy(d *schema.ResourceData) (*v1pb.ReviewRulePolicy, error) {
	return expandReviewRulePolicy(d.Get("review_rule_policy"))
}

func expandReviewRulePolicy(raw interface{}) (*v1pb.ReviewRulePolicy, error) {
	rawList, ok := raw.([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.New("review_rule_policy must be set for REVIEW_RULE")
	}
	block, ok := rawList[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid review_rule_policy")
	}
	policy := &v1pb.ReviewRulePolicy{}
	seen := map[v1pb.ReviewRuleType]bool{}
	for _, rawRule := range block["rules"].([]interface{}) {
		ruleName := rawRule.(string)
		ruleValue, ok := v1pb.ReviewRuleType_value[ruleName]
		if !ok || ruleValue == int32(v1pb.ReviewRuleType_REVIEW_RULE_TYPE_UNSPECIFIED) {
			return nil, errors.Errorf("invalid review rule %q", ruleName)
		}
		rule := v1pb.ReviewRuleType(ruleValue)
		if seen[rule] {
			return nil, errors.Errorf("review rule %s is listed twice", ruleName)
		}
		seen[rule] = true
		policy.Rules = append(policy.Rules, rule)
	}
	return policy, nil
}

func flattenReviewRulePolicy(p *v1pb.ReviewRulePolicy) []interface{} {
	rules := []string{}
	for _, rule := range p.Rules {
		rules = append(rules, rule.String())
	}
	return []interface{}{map[string]interface{}{"rules": rules}}
}

func validateReviewAIPolicyContent(content string) error {
	if strings.TrimSpace(content) == "" {
		return errors.New("AI review policy content must not be blank")
	}
	if len(content) > maxReviewAIPolicyBytes {
		return errors.Errorf("AI review policy content is %d bytes, the limit is %d", len(content), maxReviewAIPolicyBytes)
	}
	return nil
}

func expandReviewAIPolicy(raw interface{}) (*v1pb.ReviewAIPolicy, error) {
	rawList, ok := raw.([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.New("review_ai_policy must be set for REVIEW_AI")
	}
	block, ok := rawList[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid review_ai_policy")
	}
	content, _ := block["content"].(string)
	if err := validateReviewAIPolicyContent(content); err != nil {
		return nil, err
	}
	return &v1pb.ReviewAIPolicy{Content: content}, nil
}

func flattenReviewAIPolicy(p *v1pb.ReviewAIPolicy) []interface{} {
	return []interface{}{map[string]interface{}{"content": p.GetContent()}}
}
