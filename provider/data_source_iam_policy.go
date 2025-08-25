package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceIAMPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "The IAM policy data source.",
		ReadContext: dataSourceIAMPolicyRead,
		Schema: map[string]*schema.Schema{
			"parent": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					// workspace policy
					fmt.Sprintf("^%s$", internal.WorkspaceName),
					// project policy
					fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern),
				),
				Description: `The IAM policy parent name for the policy, support "projects/{resource id}" or "workspaces/-"`,
			},
			"iam_policy": getIAMPolicySchema(true),
		},
	}
}

func getIAMPolicySchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed: computed,
		Optional: true,
		Default:  nil,
		Type:     schema.TypeList,
		MinItems: 0,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"binding": getIAMBindingSchema(false),
			},
		},
	}
}

func getIAMBindingSchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Computed:    computed,
		Optional:    !computed,
		Description: "The binding in the IAM policy.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"role": {
					Type:        schema.TypeString,
					Computed:    computed,
					Optional:    !computed,
					Description: "The role full name in roles/{id} format.",
					ValidateDiagFunc: internal.ResourceNameValidation(
						fmt.Sprintf("^%s", internal.RoleNamePrefix),
					),
				},
				"members": {
					Type:        schema.TypeSet,
					Computed:    computed,
					Optional:    !computed,
					Description: `A set of memebers. The value can be "allUsers", "user:{email}" or "group:{email}".`,
					Elem: &schema.Schema{
						Type: schema.TypeString,
						ValidateDiagFunc: internal.ResourceNameValidation(
							"allUsers",
							"^user:",
							"^group:",
						),
					},
				},
				"condition": {
					Type:        schema.TypeSet,
					Computed:    computed,
					Optional:    true,
					Description: "Match the condition limit.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"database": {
								Type:        schema.TypeString,
								Computed:    computed,
								Optional:    true,
								Description: "The accessible database full name in instances/{instance resource id}/databases/{database name} format",
							},
							"schema": {
								Type:        schema.TypeString,
								Computed:    computed,
								Optional:    true,
								Description: "The accessible schema in the database",
							},
							"tables": {
								Type:     schema.TypeSet,
								Computed: computed,
								Optional: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
								Set:         schema.HashString,
								Description: "The accessible table list",
							},
							"row_limit": {
								Type:        schema.TypeInt,
								Computed:    computed,
								Optional:    true,
								Description: "The export row limit for exporter role",
							},
							"expire_timestamp": {
								Type:        schema.TypeString,
								Computed:    computed,
								Optional:    true,
								Description: "The expiration timestamp in YYYY-MM-DDThh:mm:ssZ format",
							},
						},
					},
					Set: conditionHash,
				},
			},
		},
		Set: bindingHash,
	}
}

func dataSourceIAMPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	parent := d.Get("parent").(string)

	var iamPolicy *v1pb.IamPolicy

	if strings.HasPrefix(parent, internal.ProjectNamePrefix) {
		projectIAM, err := c.GetProjectIAMPolicy(ctx, parent)
		if err != nil {
			return diag.FromErr(err)
		}
		iamPolicy = projectIAM
	} else {
		workspaceIAM, err := c.GetWorkspaceIAMPolicy(ctx)
		if err != nil {
			return diag.FromErr(err)
		}
		iamPolicy = workspaceIAM
	}

	d.SetId(parent)
	return setIAMPolicyMessage(d, iamPolicy)
}

func setIAMPolicyMessage(d *schema.ResourceData, iamPolicy *v1pb.IamPolicy) diag.Diagnostics {
	flattenPolicy, err := flattenIAMPolicy(iamPolicy)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("iam_policy", flattenPolicy); err != nil {
		return diag.Errorf("cannot set iam_policy: %s", err.Error())
	}
	return nil
}

func flattenIAMPolicy(p *v1pb.IamPolicy) ([]interface{}, error) {
	bindingList := []interface{}{}
	for _, binding := range p.Bindings {
		rawBinding := map[string]interface{}{}
		rawCondition := map[string]interface{}{}
		if condition := binding.Condition; condition != nil && condition.Expression != "" {
			expressions := strings.Split(condition.Expression, " && ")
			for _, expression := range expressions {
				if strings.HasPrefix(expression, `resource.database == "`) {
					rawCondition["database"] = strings.TrimSuffix(
						strings.TrimPrefix(expression, `resource.database == "`),
						`"`,
					)
				}
				if strings.HasPrefix(expression, `resource.schema == "`) {
					rawCondition["schema"] = strings.TrimSuffix(
						strings.TrimPrefix(expression, `resource.schema == "`),
						`"`,
					)
				}
				if strings.HasPrefix(expression, `resource.table in [`) {
					tableStr := strings.TrimSuffix(
						strings.TrimPrefix(expression, `resource.table in [`),
						`]`,
					)
					rawTableList := []interface{}{}
					for _, t := range strings.Split(tableStr, ",") {
						rawTableList = append(rawTableList, strings.TrimSuffix(
							strings.TrimPrefix(t, `"`),
							`"`,
						))
					}
					rawCondition["tables"] = schema.NewSet(schema.HashString, rawTableList)
				}
				if strings.HasPrefix(expression, `request.row_limit <= `) {
					i, err := strconv.Atoi(strings.TrimPrefix(expression, `request.row_limit <= `))
					if err != nil {
						return nil, errors.Errorf("cannot convert %s to int with error: %s", expression, err.Error())
					}
					rawCondition["row_limit"] = i
				}
				if strings.HasPrefix(expression, "request.time < ") {
					rawCondition["expire_timestamp"] = strings.TrimSuffix(
						strings.TrimPrefix(expression, `request.time < timestamp("`),
						`")`,
					)
				}
			}
		}

		// Only set condition if it's not empty
		if len(rawCondition) > 0 {
			rawBinding["condition"] = schema.NewSet(conditionHash, []interface{}{rawCondition})
		}
		rawBinding["role"] = binding.Role
		// Convert members slice to a set with proper interface conversion
		membersList := make([]interface{}, len(binding.Members))
		for i, member := range binding.Members {
			membersList[i] = member
		}
		rawBinding["members"] = schema.NewSet(schema.HashString, membersList)
		bindingList = append(bindingList, rawBinding)
	}

	policy := map[string]interface{}{
		"binding": schema.NewSet(bindingHash, bindingList),
	}
	return []interface{}{policy}, nil
}

func bindingHash(rawBinding interface{}) int {
	binding, err := convertToV1Binding(rawBinding)
	if err != nil {
		return 0
	}
	return internal.ToHash(binding)
}

func conditionHash(rawCondition interface{}) int {
	condition, err := convertToV1Condition(rawCondition)
	if err != nil {
		return 0
	}
	return internal.ToHashcodeInt(condition.Expression)
}
