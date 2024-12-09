package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/type/expr"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "The policy resource.",
		CreateContext: resourcePolicyCreate,
		ReadContext:   resourcePolicyRead,
		UpdateContext: resourcePolicyUpdate,
		DeleteContext: resourcePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"parent": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					// workspace policy
					regexp.MustCompile("^$"),
					// environment policy
					regexp.MustCompile(fmt.Sprintf("^%s%s$", internal.EnvironmentNamePrefix, internal.ResourceIDPattern)),
					// instance policy
					regexp.MustCompile(fmt.Sprintf("^%s%s$", internal.InstanceNamePrefix, internal.ResourceIDPattern)),
					// project policy
					regexp.MustCompile(fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern)),
					// database policy
					regexp.MustCompile(fmt.Sprintf("^%s%s/%s%s$", internal.InstanceNamePrefix, internal.ResourceIDPattern, internal.DatabaseIDPrefix, internal.ResourceIDPattern)),
				),
				Description: "The policy parent name for the policy, support projects/{resource id}, environments/{resource id}, instances/{resource id}, or instances/{resource id}/databases/{database name}",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					v1pb.PolicyType_MASKING.String(),
					v1pb.PolicyType_MASKING_EXCEPTION.String(),
				}, false),
				Description: "The policy type.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy full name",
			},
			"enforce": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Decide if the policy is enforced.",
			},
			"inherit_from_parent": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Decide if the policy should inherit from the parent.",
			},
			"masking_policy":           getMaskingPolicySchema(false),
			"masking_exception_policy": getMaskingExceptionPolicySchema(false),
		},
	}
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	parent := fmt.Sprintf("%s/%s%s", d.Get("parent").(string), internal.PolicyNamePrefix, d.Get("type").(string))
	policy, err := c.GetPolicy(ctx, parent)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policy.Name)
	return setPolicyMessage(d, policy)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	policyName := d.Id()
	if err := c.DeletePolicy(ctx, policyName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	policyName := fmt.Sprintf("%s/%s%s", d.Get("parent").(string), internal.PolicyNamePrefix, d.Get("type").(string))

	inheritFromParent := d.Get("inherit_from_parent").(bool)
	enforce := d.Get("enforce").(bool)

	_, policyType, err := internal.GetPolicyParentAndType(policyName)
	if err != nil {
		return diag.FromErr(err)
	}

	patch := &v1pb.Policy{
		Name:              policyName,
		InheritFromParent: inheritFromParent,
		Enforce:           enforce,
		Type:              policyType,
	}

	switch policyType {
	case v1pb.PolicyType_MASKING:
		maskingPolicy, err := convertToMaskingPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.Policy = &v1pb.Policy_MaskingPolicy{
			MaskingPolicy: maskingPolicy,
		}
	case v1pb.PolicyType_MASKING_EXCEPTION:
		maskingExceptionPolicy, err := convertToMaskingExceptionPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.Policy = &v1pb.Policy_MaskingExceptionPolicy{
			MaskingExceptionPolicy: maskingExceptionPolicy,
		}
	}

	var diags diag.Diagnostics
	p, err := c.UpsertPolicy(ctx, patch, []string{"inherit_from_parent", "enforce", "payload"})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to upsert policy",
			Detail:   fmt.Sprintf("Upsert policy %s failed, error: %v", policyName, err),
		})
		return diags
	}

	d.SetId(p.Name)

	diag := resourcePolicyRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	policyName := d.Id()
	_, policyType, err := internal.GetPolicyParentAndType(policyName)
	if err != nil {
		return diag.FromErr(err)
	}

	patch := &v1pb.Policy{
		Name:              policyName,
		InheritFromParent: d.Get("inherit_from_parent").(bool),
		Enforce:           d.Get("enforce").(bool),
		Type:              policyType,
	}

	updateMasks := []string{}
	if d.HasChange("inherit_from_parent") {
		updateMasks = append(updateMasks, "inherit_from_parent")
	}
	if d.HasChange("enforce") {
		updateMasks = append(updateMasks, "enforce")
	}

	if d.HasChange("masking_policy") {
		updateMasks = append(updateMasks, "payload")
		maskingPolicy, err := convertToMaskingPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.Policy = &v1pb.Policy_MaskingPolicy{
			MaskingPolicy: maskingPolicy,
		}
	}
	if d.HasChange("masking_exception_policy") {
		updateMasks = append(updateMasks, "payload")
		maskingExceptionPolicy, err := convertToMaskingExceptionPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.Policy = &v1pb.Policy_MaskingExceptionPolicy{
			MaskingExceptionPolicy: maskingExceptionPolicy,
		}
	}

	var diags diag.Diagnostics
	if len(updateMasks) > 0 {
		if _, err := c.UpsertPolicy(ctx, patch, updateMasks); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to upsert policy",
				Detail:   fmt.Sprintf("Upsert policy %s failed, error: %v", policyName, err),
			})
			return diags
		}
	}

	diag := resourcePolicyRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func convertToMaskingLevel(level string) v1pb.MaskingLevel {
	return v1pb.MaskingLevel(v1pb.MaskingLevel_value[level])
}

func convertToMaskingExceptionPolicy(d *schema.ResourceData) (*v1pb.MaskingExceptionPolicy, error) {
	rawList, ok := d.Get("masking_exception_policy").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid masking_exception_policy")
	}

	raw := rawList[0].(map[string]interface{})
	exceptionList := raw["exceptions"].([]interface{})

	policy := &v1pb.MaskingExceptionPolicy{}

	for _, exception := range exceptionList {
		rawException := exception.(map[string]interface{})

		databaseFullName := rawException["database"].(string)
		instanceID, databaseName, err := internal.GetInstanceDatabaseID(databaseFullName)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid database full name: %v", databaseFullName)
		}

		expressions := []string{
			fmt.Sprintf(`resource.instance_id == "%s"`, instanceID),
			fmt.Sprintf(`resource.database_name == "%s"`, databaseName),
		}
		if schema, ok := rawException["schema"].(string); ok && schema != "" {
			expressions = append(expressions, fmt.Sprintf(`resource.schema_name == "%s"`, schema))
		}
		if table, ok := rawException["table"].(string); ok && table != "" {
			expressions = append(expressions, fmt.Sprintf(`resource.table_name == "%s"`, table))
		}
		if column, ok := rawException["column"].(string); ok && column != "" {
			expressions = append(expressions, fmt.Sprintf(`resource.column_name == "%s"`, column))
		}
		if expire, ok := rawException["expire_timestamp"].(string); ok && expire != "" {
			expressions = append(expressions, fmt.Sprintf(`request.time < timestamp("%s")`, expire))
		}
		member := rawException["member"].(string)
		if !strings.HasPrefix(member, "user:") {
			return nil, errors.Errorf("member should in user:{email} format")
		}
		policy.MaskingExceptions = append(policy.MaskingExceptions, &v1pb.MaskingExceptionPolicy_MaskingException{
			Member: rawException["member"].(string),
			Action: v1pb.MaskingExceptionPolicy_MaskingException_Action(
				v1pb.MaskingExceptionPolicy_MaskingException_Action_value[rawException["action"].(string)],
			),
			MaskingLevel: convertToMaskingLevel(rawException["masking_level"].(string)),
			Condition: &expr.Expr{
				Expression: strings.Join(expressions, " && "),
			},
		})
	}
	return policy, nil
}

func convertToMaskingPolicy(d *schema.ResourceData) (*v1pb.MaskingPolicy, error) {
	rawList, ok := d.Get("masking_policy").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid masking_policy")
	}

	raw := rawList[0].(map[string]interface{})
	rawMaskList := raw["mask_data"].([]interface{})

	policy := &v1pb.MaskingPolicy{}

	for _, maskData := range rawMaskList {
		rawMask := maskData.(map[string]interface{})
		policy.MaskData = append(policy.MaskData, &v1pb.MaskData{
			Schema:                    rawMask["schema"].(string),
			Table:                     rawMask["table"].(string),
			Column:                    rawMask["column"].(string),
			FullMaskingAlgorithmId:    rawMask["full_masking_algorithm_id"].(string),
			PartialMaskingAlgorithmId: rawMask["partial_masking_algorithm_id"].(string),
			MaskingLevel:              convertToMaskingLevel(rawMask["masking_level"].(string)),
		})
	}

	return policy, nil
}
