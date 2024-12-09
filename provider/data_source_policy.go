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

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourcePolicy() *schema.Resource {
	return &schema.Resource{
		Description: "The policy data source.",
		ReadContext: dataSourcePolicyRead,
		Schema: map[string]*schema.Schema{
			"parent": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
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
			"inherit_from_parent": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Decide if the policy should inherit from the parent.",
			},
			"enforce": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Decide if the policy is enforced.",
			},
			"masking_policy":           getMaskingPolicySchema(true),
			"masking_exception_policy": getMaskingExceptionPolicySchema(true),
		},
	}
}

func getMaskingExceptionPolicySchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed: computed,
		Optional: true,
		Default:  nil,
		Type:     schema.TypeList,
		MinItems: 0,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"exceptions": {
					Computed: computed,
					Optional: true,
					Default:  nil,
					MinItems: 0,
					Type:     schema.TypeList,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"database": {
								Type:         schema.TypeString,
								Computed:     computed,
								Optional:     true,
								ValidateFunc: validation.StringIsNotEmpty,
								Description:  "The database full name in instances/{instance resource id}/databases/{database name} format",
							},
							"schema": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
							},
							"table": {
								Type:         schema.TypeString,
								Computed:     computed,
								Optional:     true,
								ValidateFunc: validation.StringIsNotEmpty,
							},
							"column": {
								Type:         schema.TypeString,
								Computed:     computed,
								Optional:     true,
								ValidateFunc: validation.StringIsNotEmpty,
							},
							"member": {
								Type:         schema.TypeString,
								Computed:     computed,
								Optional:     true,
								ValidateFunc: validation.StringIsNotEmpty,
								Description:  "The member in user:{email} format.",
							},
							"masking_level": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
								ValidateFunc: validation.StringInSlice([]string{
									v1pb.MaskingLevel_NONE.String(),
									v1pb.MaskingLevel_PARTIAL.String(),
								}, false),
							},
							"action": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
								ValidateFunc: validation.StringInSlice([]string{
									v1pb.MaskingExceptionPolicy_MaskingException_QUERY.String(),
									v1pb.MaskingExceptionPolicy_MaskingException_EXPORT.String(),
								}, false),
							},
							"expire_timestamp": {
								Type:        schema.TypeString,
								Computed:    computed,
								Optional:    true,
								Description: "The expiration timestamp in YYYY-MM-DDThh:mm:ss.000Z format",
							},
						},
					},
				},
			},
		},
	}
}

func getMaskingPolicySchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed: computed,
		Optional: true,
		Default:  nil,
		Type:     schema.TypeList,
		MinItems: 0,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"mask_data": {
					MinItems: 0,
					Computed: computed,
					Optional: true,
					Default:  nil,
					Type:     schema.TypeList,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"schema": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
							},
							"table": {
								Type:         schema.TypeString,
								Computed:     computed,
								Optional:     true,
								ValidateFunc: validation.StringIsNotEmpty,
							},
							"column": {
								Type:         schema.TypeString,
								Computed:     computed,
								Optional:     true,
								ValidateFunc: validation.StringIsNotEmpty,
							},
							"full_masking_algorithm_id": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
							},
							"partial_masking_algorithm_id": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
							},
							"masking_level": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
								ValidateFunc: validation.StringInSlice([]string{
									v1pb.MaskingLevel_NONE.String(),
									v1pb.MaskingLevel_PARTIAL.String(),
									v1pb.MaskingLevel_FULL.String(),
								}, false),
							},
						},
					},
				},
			},
		},
	}
}

func dataSourcePolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	policyName := fmt.Sprintf("%s/%s%s", d.Get("parent").(string), internal.PolicyNamePrefix, d.Get("type").(string))
	policy, err := c.GetPolicy(ctx, policyName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policy.Name)
	return setPolicyMessage(d, policy)
}

func setPolicyMessage(d *schema.ResourceData, policy *v1pb.Policy) diag.Diagnostics {
	parent, _, err := internal.GetPolicyParentAndType(policy.Name)
	if err != nil {
		return diag.Errorf("cannot parse name for policy: %s", err.Error())
	}
	if err := d.Set("name", policy.Name); err != nil {
		return diag.Errorf("cannot set name for policy: %s", err.Error())
	}
	if err := d.Set("parent", parent); err != nil {
		return diag.Errorf("cannot set parent for policy: %s", err.Error())
	}
	if err := d.Set("inherit_from_parent", policy.InheritFromParent); err != nil {
		return diag.Errorf("cannot set inherit_from_parent for policy: %s", err.Error())
	}
	if err := d.Set("enforce", policy.Enforce); err != nil {
		return diag.Errorf("cannot set enforce for policy: %s", err.Error())
	}

	if p := policy.GetMaskingPolicy(); p != nil {
		if err := d.Set("masking_policy", flattenMaskingPolicy(p)); err != nil {
			return diag.Errorf("cannot set masking_policy: %s", err.Error())
		}
	}
	if p := policy.GetMaskingExceptionPolicy(); p != nil {
		exceptionPolicy, err := flattenMaskingExceptionPolicy(p)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("masking_exception_policy", exceptionPolicy); err != nil {
			return diag.Errorf("cannot set masking_policy: %s", err.Error())
		}
	}

	return nil
}

func flattenMaskingPolicy(p *v1pb.MaskingPolicy) []interface{} {
	maskDataList := []interface{}{}
	for _, maskData := range p.MaskData {
		raw := map[string]interface{}{}
		raw["schema"] = maskData.Schema
		raw["table"] = maskData.Table
		raw["column"] = maskData.Column
		raw["full_masking_algorithm_id"] = maskData.FullMaskingAlgorithmId
		raw["partial_masking_algorithm_id"] = maskData.PartialMaskingAlgorithmId
		raw["masking_level"] = maskData.MaskingLevel.String()
		maskDataList = append(maskDataList, raw)
	}
	policy := map[string]interface{}{
		"mask_data": maskDataList,
	}
	return []interface{}{policy}
}

func flattenMaskingExceptionPolicy(p *v1pb.MaskingExceptionPolicy) ([]interface{}, error) {
	exceptionList := []interface{}{}
	for _, exception := range p.MaskingExceptions {
		raw := map[string]interface{}{}
		raw["member"] = exception.Member
		raw["action"] = exception.Action.String()
		raw["masking_level"] = exception.MaskingLevel.String()

		expressions := strings.Split(exception.Condition.Expression, " && ")
		instanceID := ""
		databaseName := ""
		for _, expression := range expressions {
			if strings.HasPrefix(expression, "resource.instance_id == ") {
				instanceID = strings.TrimSuffix(
					strings.TrimPrefix(expression, `resource.instance_id == "`),
					`"`,
				)
			}
			if strings.HasPrefix(expression, "resource.database_name == ") {
				databaseName = strings.TrimSuffix(
					strings.TrimPrefix(expression, `resource.database_name == "`),
					`"`,
				)
			}
			if strings.HasPrefix(expression, "resource.table_name == ") {
				raw["table"] = strings.TrimSuffix(
					strings.TrimPrefix(expression, `resource.table_name == "`),
					`"`,
				)
			}
			if strings.HasPrefix(expression, "resource.schema_name == ") {
				raw["schema"] = strings.TrimSuffix(
					strings.TrimPrefix(expression, `resource.schema_name == "`),
					`"`,
				)
			}
			if strings.HasPrefix(expression, "resource.column_name == ") {
				raw["column"] = strings.TrimSuffix(
					strings.TrimPrefix(expression, `resource.column_name == "`),
					`"`,
				)
			}
			if strings.HasPrefix(expression, "request.time < ") {
				raw["expire_timestamp"] = strings.TrimSuffix(
					strings.TrimPrefix(expression, `request.time < timestamp("`),
					`")`,
				)
			}
		}
		if instanceID == "" || databaseName == "" {
			return nil, errors.Errorf("invalid exception policy condition: %v", exception.Condition.Expression)
		}
		raw["database"] = fmt.Sprintf("%s%s/%s%s", internal.InstanceNamePrefix, instanceID, internal.DatabaseIDPrefix, databaseName)
		exceptionList = append(exceptionList, raw)
	}
	policy := map[string]interface{}{
		"exceptions": exceptionList,
	}
	return []interface{}{policy}, nil
}
