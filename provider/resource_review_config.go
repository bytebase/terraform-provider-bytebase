package provider

import (
	"context"
	"fmt"
	"strings"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
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
				Type:        schema.TypeList,
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
								v1pb.SQLReviewRuleLevel_WARNING.String(),
								v1pb.SQLReviewRuleLevel_ERROR.String(),
								v1pb.SQLReviewRuleLevel_DISABLED.String(),
							}, false),
						},
						"payload": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The payload is a JSON string that varies by rule type. Check https://github.com/bytebase/bytebase/blob/main/proto/v1/v1/SQL_REVIEW_RULES_DOCUMENTATION.md#payload-structure-types for all details",
						},
						"comment": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The comment for the rule.",
						},
					},
				},
			},
		},
	}
}

func resourceReviewConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	fullName := d.Id()
	review, err := c.GetReviewConfig(ctx, fullName)
	if err != nil {
		return diag.FromErr(err)
	}

	return setReviewConfig(d, review)
}

func resourceReviewConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	resources := getReviewConfigRelatedResources(d)

	removeReviewConfigTag(ctx, c, resources)
	return internal.ResourceDelete(ctx, d, m)
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
	if err := d.Set("rules", flattenReviewRules(review.Rules)); err != nil {
		return diag.Errorf("cannot set rules for review: %s", err.Error())
	}

	return nil
}

func flattenReviewRules(rules []*v1pb.SQLReviewRule) []interface{} {
	ruleList := []interface{}{}
	for _, rule := range rules {
		rawRule := map[string]interface{}{}
		rawRule["type"] = rule.Type
		rawRule["engine"] = rule.Engine.String()
		rawRule["level"] = rule.Level.String()
		rawRule["comment"] = rule.Comment
		rawRule["payload"] = rule.Payload
		ruleList = append(ruleList, rawRule)
	}
	return ruleList
}

func convertToV1RuleList(d *schema.ResourceData) ([]*v1pb.SQLReviewRule, error) {
	ruleRawList, ok := d.Get("rules").([]interface{})
	if !ok || len(ruleRawList) == 0 {
		return nil, errors.Errorf("rules is required")
	}

	ruleList := []*v1pb.SQLReviewRule{}

	for _, r := range ruleRawList {
		rawRule := r.(map[string]interface{})
		payload := rawRule["payload"].(string)
		if payload == "" {
			payload = "{}"
		}
		ruleList = append(ruleList, &v1pb.SQLReviewRule{
			Type:    rawRule["type"].(string),
			Comment: rawRule["comment"].(string),
			Payload: payload,
			Engine:  v1pb.Engine(v1pb.Engine_value[rawRule["engine"].(string)]),
			Level:   v1pb.SQLReviewRuleLevel(v1pb.SQLReviewRuleLevel_value[rawRule["level"].(string)]),
		})
	}
	return ruleList, nil
}
