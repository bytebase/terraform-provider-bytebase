package provider

import (
	"context"
	"fmt"
	"strings"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceReviewConfig() *schema.Resource {
	return &schema.Resource{
		Description:   "The review config resource.",
		ReadContext:   resourceReviewConfigRead,
		DeleteContext: resourceReviewConfigDelete,
		CreateContext: resourceReviewConfigUpsert,
		UpdateContext: resourceReviewConfigUpsert,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The unique resource id for the review config.",
			},
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The title for the review config.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Enable the SQL review config",
			},
			"resources": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Resources using the config. We support attach the review config for environments or projects with format {resurce}/{resource id}. For example, environments/test, projects/sample.",
			},
			"rules": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Description: "The SQL review rules.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The rule unique type. Check https://github.com/bytebase/bytebase/blob/main/proto/v1/v1/SQL_REVIEW_RULES_DOCUMENTATION.md#rule-categories for all rules",
						},
						"engine": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The rule for the database engine.",
							ValidateFunc: internal.EngineValidation,
						},
						"level": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The rule level.",
							ValidateFunc: validation.StringInSlice([]string{
								v1pb.SQLReviewRule_WARNING.String(),
								v1pb.SQLReviewRule_ERROR.String(),
							}, false),
						},
						// Typed payload fields - use one based on rule type
						"number_payload": {
							Type:     schema.TypeInt,
							Optional: true,
							Description: "Number payload for rules: STATEMENT_INSERT_ROW_LIMIT, STATEMENT_AFFECTED_ROW_LIMIT, " +
								"STATEMENT_WHERE_MAXIMUM_LOGICAL_OPERATOR_COUNT, STATEMENT_MAXIMUM_LIMIT_VALUE, STATEMENT_MAXIMUM_JOIN_TABLE_COUNT, " +
								"STATEMENT_MAXIMUM_STATEMENTS_IN_TRANSACTION, COLUMN_MAXIMUM_CHARACTER_LENGTH, COLUMN_MAXIMUM_VARCHAR_LENGTH, " +
								"COLUMN_AUTO_INCREMENT_INITIAL_VALUE, INDEX_KEY_NUMBER_LIMIT, INDEX_TOTAL_NUMBER_LIMIT, " +
								"TABLE_TEXT_FIELDS_TOTAL_LENGTH, TABLE_LIMIT_SIZE, SYSTEM_COMMENT_LENGTH, ADVICE_ONLINE_MIGRATION.",
						},
						"string_payload": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "String payload for rule: STATEMENT_QUERY_MINIMUM_PLAN_LEVEL.",
						},
						"string_array_payload": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Description: "String array payload for rules: COLUMN_REQUIRED, COLUMN_TYPE_DISALLOW_LIST, INDEX_PRIMARY_KEY_TYPE_ALLOWLIST, " +
								"INDEX_TYPE_ALLOW_LIST, SYSTEM_CHARSET_ALLOWLIST, SYSTEM_COLLATION_ALLOWLIST, " +
								"SYSTEM_FUNCTION_DISALLOWED_LIST, TABLE_DISALLOW_DDL, TABLE_DISALLOW_DML.",
						},
						"naming_payload": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Description: "Naming payload for rules: NAMING_TABLE, NAMING_COLUMN, NAMING_COLUMN_AUTO_INCREMENT, " +
								"NAMING_INDEX_FK, NAMING_INDEX_IDX, NAMING_INDEX_UK, NAMING_INDEX_PK, TABLE_DROP_NAMING_CONVENTION.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"format": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The naming format regex pattern.",
									},
									"max_length": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "The maximum length for the name.",
									},
								},
							},
						},
						"comment_convention_payload": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Comment convention payload for rules: COLUMN_COMMENT, TABLE_COMMENT.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"required": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Whether the comment is required.",
									},
									"max_length": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "The maximum length for the comment.",
									},
								},
							},
						},
						"naming_case_payload": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Naming case payload for rule: NAMING_IDENTIFIER_CASE. Set to true for UPPER case, false for LOWER case.",
						},
					},
				},
				Set: reviewRuleHash,
			},
		},
	}
}

func resourceReviewConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	fullName := d.Id()
	review, err := c.GetReviewConfig(ctx, fullName)
	if err != nil {
		// Check if the resource was deleted outside of Terraform
		if internal.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Resource %s not found, removing from state", fullName))
			// Remove from state to trigger recreation on next apply
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return setReviewConfig(d, review)
}

func resourceReviewConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	resources := getReviewConfigRelatedResources(d)

	removeReviewConfigTag(ctx, c, resources)
	return internal.ResourceDelete(ctx, d, c.DeleteReviewConfig)
}

func resourceReviewConfigUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	existedName := d.Id()

	reviewID := d.Get("resource_id").(string)
	reviewName := fmt.Sprintf("%s%s", internal.ReviewConfigNamePrefix, reviewID)

	if existedName != "" {
		if existedName != reviewName {
			return diag.Errorf("cannot change the resource id")
		}
	}

	rules, err := convertToV1RuleList(d)
	if err != nil {
		return diag.FromErr(err)
	}

	reviewConfig := &v1pb.ReviewConfig{
		Name:    reviewName,
		Title:   d.Get("title").(string),
		Enabled: d.Get("enabled").(bool),
		Rules:   rules,
	}
	review, err := c.UpsertReviewConfig(ctx, reviewConfig, []string{
		"title",
		"enabled",
		"rules",
	})
	if err != nil {
		return diag.FromErr(err)
	}

	rawConfig := d.GetRawConfig()
	if config := rawConfig.GetAttr("resources"); !config.IsNull() {
		// only update attached resources if users set this field.
		pendingDelete, pendingAdd := getResourceDiff(ctx, c, d)
		removeReviewConfigTag(ctx, c, pendingDelete)
		patchTagPolicy(ctx, c, review.Name, pendingAdd)
	}

	d.SetId(review.Name)

	return resourceReviewConfigRead(ctx, d, m)
}

// getResourceDiff returns pending delete list and pending add list.
func getResourceDiff(ctx context.Context, client api.Client, d *schema.ResourceData) ([]string, []string) {
	existedName := d.Id()
	oldAttachedResources := []string{}
	if existedName != "" {
		existedReview, err := client.GetReviewConfig(ctx, existedName)
		if err != nil {
			tflog.Debug(ctx, fmt.Sprintf("get review config %s failed with error: %v", existedName, err))
		} else if existedReview != nil {
			oldAttachedResources = existedReview.Resources
		}
	}

	oldResourceMap := map[string]bool{}
	for _, resource := range oldAttachedResources {
		oldResourceMap[resource] = true
	}

	newAttachedResources := getReviewConfigRelatedResources(d)
	newResourceMap := map[string]bool{}
	for _, resource := range newAttachedResources {
		newResourceMap[resource] = true
	}

	pendingDelete := []string{}
	pendingAdd := []string{}
	for old := range oldResourceMap {
		if !newResourceMap[old] {
			pendingDelete = append(pendingDelete, old)
		}
	}
	for new := range newResourceMap {
		if !oldResourceMap[new] {
			pendingAdd = append(pendingAdd, new)
		}
	}
	return pendingDelete, pendingAdd
}

func getReviewConfigRelatedResources(d *schema.ResourceData) []string {
	resources := []string{}
	rawSet, ok := d.Get("resources").(*schema.Set)
	if !ok || rawSet.Len() == 0 {
		return resources
	}
	for _, raw := range rawSet.List() {
		resources = append(resources, raw.(string))
	}
	return resources
}

func removeReviewConfigTag(ctx context.Context, client api.Client, resources []string) {
	for _, resource := range resources {
		policyName := fmt.Sprintf("%s/%s%s", resource, internal.PolicyNamePrefix, v1pb.PolicyType_TAG.String())
		if err := client.DeletePolicy(ctx, policyName); err != nil {
			tflog.Error(ctx, fmt.Sprintf("failed to delete %v policy with error: %v", policyName, err.Error()))
		}
	}
}

func patchTagPolicy(ctx context.Context, client api.Client, reviewName string, resources []string) diag.Diagnostics {
	for _, resource := range resources {
		if !strings.HasPrefix(resource, internal.ProjectNamePrefix) && !strings.HasPrefix(resource, internal.EnvironmentNamePrefix) {
			return diag.Errorf("invalid resource, only support projects/{id} or environments/{id}")
		}
		policyName := fmt.Sprintf("%s/%s%s", resource, internal.PolicyNamePrefix, v1pb.PolicyType_TAG.String())
		if _, err := client.UpsertPolicy(ctx, &v1pb.Policy{
			Name:    policyName,
			Enforce: true,
			Type:    v1pb.PolicyType_TAG,
			Policy: &v1pb.Policy_TagPolicy{
				TagPolicy: &v1pb.TagPolicy{
					Tags: map[string]string{
						"bb.tag.review_config": reviewName,
					},
				},
			},
		}, []string{"payload", "enforce"}); err != nil {
			return diag.Errorf("failed to attach review config %s to resource %s with error: %v", reviewName, resource, err.Error())
		}
	}

	return nil
}

func setReviewConfig(d *schema.ResourceData, review *v1pb.ReviewConfig) diag.Diagnostics {
	reviewID, err := internal.GetReviewConfigID(review.Name)
	if err != nil {
		return diag.Errorf("failed to parse id from review name %s with error: %v", review.Name, err.Error())
	}

	if err := d.Set("resource_id", reviewID); err != nil {
		return diag.Errorf("cannot set resource_id for review: %s", err.Error())
	}
	if err := d.Set("title", review.Title); err != nil {
		return diag.Errorf("cannot set title for review: %s", err.Error())
	}
	if err := d.Set("enabled", review.Enabled); err != nil {
		return diag.Errorf("cannot set enabled for review: %s", err.Error())
	}
	if err := d.Set("resources", review.Resources); err != nil {
		return diag.Errorf("cannot set resources for review: %s", err.Error())
	}
	rules := flattenReviewRules(review.Rules)
	if err := d.Set("rules", schema.NewSet(reviewRuleHash, rules)); err != nil {
		return diag.Errorf("cannot set rules for review: %s", err.Error())
	}

	return nil
}

func flattenReviewRules(rules []*v1pb.SQLReviewRule) []interface{} {
	ruleList := []interface{}{}
	for _, rule := range rules {
		rawRule := map[string]interface{}{}
		rawRule["type"] = rule.Type.String()
		rawRule["engine"] = rule.Engine.String()
		rawRule["level"] = rule.Level.String()
		// Set typed payload fields based on rule payload type
		flattenRulePayload(rule, rawRule)
		ruleList = append(ruleList, rawRule)
	}
	return ruleList
}

func flattenRulePayload(rule *v1pb.SQLReviewRule, rawRule map[string]interface{}) {
	if payload := rule.GetNumberPayload(); payload != nil {
		rawRule["number_payload"] = int(payload.GetNumber())
		return
	}
	if payload := rule.GetStringPayload(); payload != nil {
		rawRule["string_payload"] = payload.GetValue()
		return
	}
	if payload := rule.GetStringArrayPayload(); payload != nil {
		rawRule["string_array_payload"] = payload.GetList()
		return
	}
	if payload := rule.GetNamingPayload(); payload != nil {
		rawRule["naming_payload"] = []interface{}{
			map[string]interface{}{
				"format":     payload.GetFormat(),
				"max_length": int(payload.GetMaxLength()),
			},
		}
		return
	}
	if payload := rule.GetCommentConventionPayload(); payload != nil {
		rawRule["comment_convention_payload"] = []interface{}{
			map[string]interface{}{
				"required":   payload.GetRequired(),
				"max_length": int(payload.GetMaxLength()),
			},
		}
		return
	}
	if payload := rule.GetNamingCasePayload(); payload != nil {
		rawRule["naming_case_payload"] = payload.GetUpper()
		return
	}
}

func convertToV1Rule(rawSchema interface{}, rawConfig cty.Value) *v1pb.SQLReviewRule {
	rawRule := rawSchema.(map[string]interface{})
	ruleType := v1pb.SQLReviewRule_Type(v1pb.SQLReviewRule_Type_value[rawRule["type"].(string)])
	rule := &v1pb.SQLReviewRule{
		Type:   ruleType,
		Engine: v1pb.Engine(v1pb.Engine_value[rawRule["engine"].(string)]),
		Level:  v1pb.SQLReviewRule_Level(v1pb.SQLReviewRule_Level_value[rawRule["level"].(string)]),
	}
	// Set payload from typed fields
	setRulePayload(rule, rawRule, rawConfig)
	return rule
}

// isAttrSet checks if an attribute was explicitly set in the raw config.
func isAttrSet(rawConfig cty.Value, attrName string) bool {
	// Check if rawConfig is valid (not cty.NilVal) and not null
	if rawConfig == cty.NilVal || rawConfig.IsNull() {
		return false
	}
	if !rawConfig.Type().HasAttribute(attrName) {
		return false
	}
	attr := rawConfig.GetAttr(attrName)
	return !attr.IsNull()
}

func setRulePayload(rule *v1pb.SQLReviewRule, rawRule map[string]interface{}, rawConfig cty.Value) {
	// Check each typed payload field and set the appropriate payload
	// Use rawConfig (via isAttrSet) to check if the field was explicitly configured

	// NumberPayload
	if isAttrSet(rawConfig, "number_payload") {
		if num, ok := rawRule["number_payload"].(int); ok {
			rule.Payload = &v1pb.SQLReviewRule_NumberPayload{
				NumberPayload: &v1pb.SQLReviewRule_NumberRulePayload{
					Number: int32(num),
				},
			}
			return
		}
	}

	// StringPayload
	if isAttrSet(rawConfig, "string_payload") {
		if str, ok := rawRule["string_payload"].(string); ok {
			rule.Payload = &v1pb.SQLReviewRule_StringPayload{
				StringPayload: &v1pb.SQLReviewRule_StringRulePayload{
					Value: str,
				},
			}
			return
		}
	}

	// StringArrayPayload
	if isAttrSet(rawConfig, "string_array_payload") {
		if list, ok := rawRule["string_array_payload"].([]interface{}); ok && len(list) > 0 {
			strList := make([]string, 0, len(list))
			for _, item := range list {
				if s, ok := item.(string); ok {
					strList = append(strList, s)
				}
			}
			rule.Payload = &v1pb.SQLReviewRule_StringArrayPayload{
				StringArrayPayload: &v1pb.SQLReviewRule_StringArrayRulePayload{
					List: strList,
				},
			}
			return
		}
	}

	// NamingPayload
	if isAttrSet(rawConfig, "naming_payload") {
		if list, ok := rawRule["naming_payload"].([]interface{}); ok && len(list) > 0 {
			if raw, ok := list[0].(map[string]interface{}); ok {
				namingPayload := &v1pb.SQLReviewRule_NamingRulePayload{}
				if format, ok := raw["format"].(string); ok {
					namingPayload.Format = format
				}
				if maxLen, ok := raw["max_length"].(int); ok {
					namingPayload.MaxLength = int32(maxLen)
				}
				rule.Payload = &v1pb.SQLReviewRule_NamingPayload{
					NamingPayload: namingPayload,
				}
				return
			}
		}
	}

	// CommentConventionPayload
	if isAttrSet(rawConfig, "comment_convention_payload") {
		if list, ok := rawRule["comment_convention_payload"].([]interface{}); ok && len(list) > 0 {
			if raw, ok := list[0].(map[string]interface{}); ok {
				commentPayload := &v1pb.SQLReviewRule_CommentConventionRulePayload{}
				if required, ok := raw["required"].(bool); ok {
					commentPayload.Required = required
				}
				if maxLen, ok := raw["max_length"].(int); ok {
					commentPayload.MaxLength = int32(maxLen)
				}
				rule.Payload = &v1pb.SQLReviewRule_CommentConventionPayload{
					CommentConventionPayload: commentPayload,
				}
				return
			}
		}
	}

	// NamingCasePayload
	if isAttrSet(rawConfig, "naming_case_payload") {
		if upper, ok := rawRule["naming_case_payload"].(bool); ok {
			rule.Payload = &v1pb.SQLReviewRule_NamingCasePayload{
				NamingCasePayload: &v1pb.SQLReviewRule_NamingCaseRulePayload{
					Upper: upper,
				},
			}
			return
		}
	}

	rule.Payload = nil
}

func convertToV1RuleList(d *schema.ResourceData) ([]*v1pb.SQLReviewRule, error) {
	ruleRawList, ok := d.Get("rules").(*schema.Set)
	if !ok || ruleRawList.Len() == 0 {
		return nil, errors.Errorf("rules is required")
	}

	// Get raw config for rules to check which attributes were explicitly set
	rawConfig := d.GetRawConfig()
	rulesConfig := rawConfig.GetAttr("rules")

	// Build a map of (type:engine) -> raw config for matching
	rawConfigMap := make(map[string]cty.Value)
	if !rulesConfig.IsNull() && rulesConfig.CanIterateElements() {
		for it := rulesConfig.ElementIterator(); it.Next(); {
			_, ruleVal := it.Element()
			if ruleVal.IsNull() {
				continue
			}
			typeVal := ruleVal.GetAttr("type")
			engineVal := ruleVal.GetAttr("engine")
			if !typeVal.IsNull() && !engineVal.IsNull() {
				key := typeVal.AsString() + ":" + engineVal.AsString()
				rawConfigMap[key] = ruleVal
			}
		}
	}

	ruleList := []*v1pb.SQLReviewRule{}

	for _, r := range ruleRawList.List() {
		rawRule := r.(map[string]interface{})
		ruleType := rawRule["type"].(string)
		ruleEngine := rawRule["engine"].(string)
		key := ruleType + ":" + ruleEngine

		// Get the matching raw config, or use null value if not found
		ruleConfig := rawConfigMap[key]
		ruleList = append(ruleList, convertToV1Rule(r, ruleConfig))
	}
	return ruleList, nil
}

func reviewRuleHash(rawSchema interface{}) int {
	// For hashing, we use a null cty.Value since we don't have raw config here.
	// The hash is based on type, engine, level and payload values from the processed map.
	rule := convertToV1Rule(rawSchema, cty.NilVal)
	return internal.ToHash(rule)
}
