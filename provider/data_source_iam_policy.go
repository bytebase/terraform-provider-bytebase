package provider

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

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
					regexp.MustCompile("^workspaces/-$"),
					// project policy
					regexp.MustCompile(fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern)),
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
				},
				"members": {
					Type:        schema.TypeSet,
					Computed:    computed,
					Optional:    !computed,
					Description: `A set of memebers. The value can be "allUsers", "user:{email}" or "group:{email}".`,
					Elem: &schema.Schema{
						Type: schema.TypeString,
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
					Set: func(i interface{}) int {
						return internal.ToHashcodeInt(conditionHash(i))
					},
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
					rawTableList := []string{}
					for _, t := range strings.Split(tableStr, ",") {
						rawTableList = append(rawTableList, strings.TrimSuffix(
							strings.TrimPrefix(t, `"`),
							`"`,
						))
					}
					rawCondition["tables"] = rawTableList
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

		rawBinding["condition"] = schema.NewSet(func(i interface{}) int {
			return internal.ToHashcodeInt(conditionHash(i))
		}, []interface{}{rawCondition})
		rawBinding["role"] = binding.Role
		rawBinding["members"] = binding.Members
		bindingList = append(bindingList, rawBinding)
	}

	policy := map[string]interface{}{
		"binding": schema.NewSet(bindingHash, bindingList),
	}
	return []interface{}{policy}, nil
}

func bindingHash(rawBinding interface{}) int {
	var buf bytes.Buffer
	binding := rawBinding.(map[string]interface{})

	if v, ok := binding["role"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}

	if condition, ok := binding["condition"].(*schema.Set); ok && condition.Len() > 0 && condition.List()[0] != nil {
		rawCondition := condition.List()[0].(map[string]interface{})
		_, _ = buf.WriteString(conditionHash(rawCondition))
	}

	return internal.ToHashcodeInt(buf.String())
}

func conditionHash(rawCondition interface{}) string {
	var buf bytes.Buffer
	condition := rawCondition.(map[string]interface{})

	if v, ok := condition["database"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := condition["schema"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := condition["tables"].(*schema.Set); ok {
		for _, t := range v.List() {
			_, _ = buf.WriteString(fmt.Sprintf("table.%s-", t.(string)))
		}
	}
	if v, ok := condition["row_limit"].(int); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%d-", v))
	}
	if v, ok := condition["expire_timestamp"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}

	return buf.String()
}
