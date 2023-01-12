package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourcePolicyList() *schema.Resource {
	return &schema.Resource{
		Description: "The policy data source list.",
		ReadContext: dataSourcePolicyListRead,
		Schema: map[string]*schema.Schema{
			"show_deleted": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Including removed policy in the response.",
			},
			"project": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The project resource id for the policy.",
			},
			"environment": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The environment resource id for the policy.",
			},
			"instance": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The instance resource id for the policy.",
			},
			"database": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The database name for the policy.",
			},
			"policies": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy type.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy name",
						},
						"inherit_from_parent": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Decide if the policy should inherit from the parent.",
						},
						"deployment_approval_policy": deploymentApprovalPolicySchema,
						"backup_plan_policy":         backupPlanPolicySchema,
						"sensitive_data_policy":      sensitiveDataPolicy,
						"access_control_policy":      accessControlPolicy,
						"sql_review_policy":          sqlReviewPolicy,
					},
				},
			},
		},
	}
}

func dataSourcePolicyListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	find, err := getPolicyFind(d)
	if err != nil {
		return diag.FromErr(err)
	}
	find.ShowDeleted = d.Get("show_deleted").(bool)

	response, err := c.ListPolicies(ctx, find)
	if err != nil {
		return diag.FromErr(err)
	}

	policies := make([]map[string]interface{}, 0)
	for _, policy := range response.Policies {
		raw := make(map[string]interface{})
		raw["name"] = policy.Name
		raw["inherit_from_parent"] = policy.InheritFromParent
		if p := policy.DeploymentApprovalPolicy; p != nil {
			raw["deployment_approval_policy"] = flattenDeploymentApprovalPolicy(p)
		}
		if p := policy.BackupPlanPolicy; p != nil {
			raw["backup_plan_policy"] = flattenBackupPlanPolicy(p)
		}
		if p := policy.SensitiveDataPolicy; p != nil {
			raw["sensitive_data_policy"] = flattenSensitiveDataPolicy(p)
		}
		if p := policy.AccessControlPolicy; p != nil {
			raw["access_control_policy"] = flattenAccessControlPolicy(p)
		}
		if p := policy.SQLReviewPolicy; p != nil {
			raw["sql_review_policy"] = flattenSQLReviewPolicy(p)
		}
		policies = append(policies, raw)
	}

	if err := d.Set("policies", policies); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
