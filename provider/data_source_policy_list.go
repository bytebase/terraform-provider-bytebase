package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"

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
				Default:  internal.WorkspaceName,
				ValidateDiagFunc: internal.ResourceNameValidation(
					// workspace policy
					fmt.Sprintf("^%s$", internal.WorkspaceName),
					// environment policy
					fmt.Sprintf("^%s%s$", internal.EnvironmentNamePrefix, internal.ResourceIDPattern),
					// instance policy
					fmt.Sprintf("^%s%s$", internal.InstanceNamePrefix, internal.ResourceIDPattern),
					// project policy
					fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern),
					// database policy
					fmt.Sprintf(`^%s%s/%s\S+$`, internal.InstanceNamePrefix, internal.ResourceIDPattern, internal.DatabaseIDPrefix),
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
						"disable_copy_data_policy": getDisableCopyDataPolicySchema(true),
						"data_source_query_policy": getDataSourceQueryPolicySchema(true),
						"rollout_policy":           getRolloutPolicySchema(true),
						"query_data_policy":        getDataQueryPolicySchema(true),
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

	response, err := c.ListPolicies(ctx, parent)
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

		key, payload, diags := flattenPolicyPayload(policy)
		if diags != nil {
			return diags
		}
		raw[key] = payload

		policies = append(policies, raw)
	}

	if err := d.Set("policies", policies); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
