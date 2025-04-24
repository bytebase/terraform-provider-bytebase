package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourcePolicyList() *schema.Resource {
	return &schema.Resource{
		Description: "The policy data source list.",
		ReadContext: dataSourcePolicyListRead,
		Schema: map[string]*schema.Schema{
			"parent": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ValidateDiagFunc: internal.ResourceNameValidation(
					// workspace policy
					regexp.MustCompile(fmt.Sprintf("^%s$", internal.WorkspaceName)),
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
						"masking_exception_policy": getMaskingExceptionPolicySchema(true),
						"global_masking_policy":    getGlobalMaskingPolicySchema(true),
					},
				},
			},
		},
	}
}

func dataSourcePolicyListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	parent := d.Get("parent").(string)
	if parent == internal.WorkspaceName {
		parent = ""
	}

	response, err := c.ListPolicies(ctx, d.Get("parent").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	policies := make([]map[string]interface{}, 0)
	for _, policy := range response.Policies {
		if policy.Type != v1pb.PolicyType_MASKING_EXCEPTION && policy.Type != v1pb.PolicyType_MASKING_RULE {
			continue
		}
		raw := make(map[string]interface{})
		raw["name"] = policy.Name
		raw["type"] = policy.Type.String()
		raw["inherit_from_parent"] = policy.InheritFromParent
		raw["enforce"] = policy.Enforce

		if p := policy.GetMaskingExceptionPolicy(); p != nil {
			exceptionPolicy, err := flattenMaskingExceptionPolicy(p)
			if err != nil {
				return diag.FromErr(err)
			}
			raw["masking_exception_policy"] = exceptionPolicy
		}
		if p := policy.GetMaskingRulePolicy(); p != nil {
			maskingPolicy, err := flattenGlobalMaskingPolicy(p)
			if err != nil {
				return diag.FromErr(err)
			}
			raw["global_masking_policy"] = maskingPolicy
		}

		policies = append(policies, raw)
	}

	if err := d.Set("policies", policies); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
