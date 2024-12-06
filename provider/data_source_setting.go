package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	v1alpha1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceSetting() *schema.Resource {
	return &schema.Resource{
		Description: "The setting data source.",
		ReadContext: dataSourceSettingRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(api.SettingWorkspaceApproval),
					string(api.SettingWorkspaceExternalApproval),
				}, false),
			},
			"value": {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"string": {
							Optional: true,
							Computed: true,
							Default:  nil,
							Type:     schema.TypeString,
						},
						"workspace_approval_setting": getWorkspaceApprovalSetting(true),
					},
				},
			},
		},
	}
}

func getWorkspaceApprovalSetting(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed: computed,
		Optional: true,
		Default:  nil,
		Type:     schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"rules": {
					Type:     schema.TypeList,
					Computed: computed,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"flow": {
								Computed: computed,
								Type:     schema.TypeList,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"title": {
											Type:     schema.TypeString,
											Computed: computed,
										},
										"description": {
											Type:     schema.TypeString,
											Computed: computed,
										},
										"creator": {
											Type:     schema.TypeString,
											Computed: computed,
										},
										"steps": {
											Type:     schema.TypeList,
											Computed: computed,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"type": {
														Type:     schema.TypeString,
														Computed: computed,
														Optional: true,
														ValidateFunc: validation.StringInSlice([]string{
															v1pb.ApprovalStep_ALL.String(),
															v1pb.ApprovalStep_ANY.String(),
														}, false),
													},
													"nodes": {
														Type:     schema.TypeList,
														Computed: computed,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"type": {
																	Type:     schema.TypeString,
																	Computed: computed,
																	Optional: true,
																	ValidateFunc: validation.StringInSlice([]string{
																		v1pb.ApprovalNode_ANY_IN_GROUP.String(),
																	}, false),
																},
																"group_value": {
																	Optional: true,
																	Default:  nil,
																	Computed: computed,
																	Type:     schema.TypeString,
																	ValidateFunc: validation.StringInSlice([]string{
																		v1pb.ApprovalNode_WORKSPACE_OWNER.String(),
																		v1pb.ApprovalNode_WORKSPACE_DBA.String(),
																		v1pb.ApprovalNode_PROJECT_OWNER.String(),
																		v1pb.ApprovalNode_PROJECT_MEMBER.String(),
																	}, false),
																},
																"role": {
																	Optional:    true,
																	Default:     nil,
																	Computed:    computed,
																	Type:        schema.TypeString,
																	Description: "role name in roles/{role} format",
																},
																"external_node_id": {
																	Optional: true,
																	Default:  nil,
																	Computed: computed,
																	Type:     schema.TypeString,
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							"conditions": {
								Computed: computed,
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"source": {
											Type:     schema.TypeString,
											Computed: computed,
											Optional: true,
											ValidateFunc: validation.StringInSlice([]string{
												v1pb.Risk_DDL.String(),
												v1pb.Risk_DML.String(),
												v1pb.Risk_CREATE_DATABASE.String(),
												v1pb.Risk_DATA_EXPORT.String(),
												v1pb.Risk_REQUEST_QUERY.String(),
												v1pb.Risk_REQUEST_EXPORT.String(),
											}, false),
										},
										"level": {
											Type:     schema.TypeString,
											Computed: computed,
											Optional: true,
											ValidateFunc: validation.StringInSlice([]string{
												string(api.RiskLevelDefault),
												string(api.RiskLevelLow),
												string(api.RiskLevelModerate),
												string(api.RiskLevelHigh),
											}, false),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceSettingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	settingName := fmt.Sprintf("%s%s", internal.SettingNamePrefix, d.Get("name").(string))
	setting, err := c.GetSetting(ctx, settingName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(setting.Name)
	return setSettingMessage(ctx, d, c, setting)
}

func setSettingMessage(ctx context.Context, d *schema.ResourceData, client api.Client, setting *v1pb.Setting) diag.Diagnostics {
	if value := setting.Value.GetWorkspaceApprovalSettingValue(); value != nil {
		settingVal, err := flattenWorkspaceApprovalSetting(ctx, client, value)
		if err != nil {
			return diag.Errorf("failed to parse workspace_approval_setting: %s", err.Error())
		}
		workspaceApprovalSetting := map[string]interface{}{
			"workspace_approval_setting": settingVal,
		}
		if err := d.Set("value", []interface{}{workspaceApprovalSetting}); err != nil {
			return diag.Errorf("cannot set workspace_approval_setting: %s", err.Error())
		}
	}

	return nil
}

func parseApprovalExpression(callExpr *v1alpha1.Expr_Call) ([]map[string]interface{}, error) {
	if callExpr == nil {
		return nil, errors.Errorf("failed to parse the expression")
	}

	switch callExpr.Function {
	case "_&&_":
		resp := map[string]interface{}{}
		for _, arg := range callExpr.Args {
			argExpr := arg.GetCallExpr()
			if argExpr == nil {
				return nil, errors.Errorf("expect call_expr")
			}
			if argExpr.Function != "_==_" {
				return nil, errors.Errorf("expect == operation but found: %v", argExpr.Function)
			}
			if len(argExpr.Args) != 2 {
				return nil, errors.Errorf("expect 2 args")
			}

			if argExpr.Args[0].GetIdentExpr() == nil || argExpr.Args[1].GetConstExpr() == nil {
				return nil, errors.Errorf("expect ident expr and const expr")
			}

			argName := argExpr.Args[0].GetIdentExpr().Name
			switch argName {
			case "source":
				resp[argName] = argExpr.Args[1].GetConstExpr().GetStringValue()
			case "level":
				levelNumber := argExpr.Args[1].GetConstExpr().GetInt64Value()
				switch levelNumber {
				case 0:
					resp[argName] = api.RiskLevelDefault
				case 100:
					resp[argName] = api.RiskLevelLow
				case 200:
					resp[argName] = api.RiskLevelModerate
				case 300:
					resp[argName] = api.RiskLevelHigh
				default:
					return nil, errors.Errorf("unknown risk level: %v", levelNumber)
				}
			default:
				return nil, errors.Errorf("unsupport arg: %v", argName)
			}
		}
		return []map[string]interface{}{resp}, nil
	case "_||_":
		resp := []map[string]interface{}{}
		for _, arg := range callExpr.Args {
			expression, err := parseApprovalExpression(arg.GetCallExpr())
			if err != nil {
				return nil, err
			}
			resp = append(resp, expression...)
		}
		return resp, nil
	default:
		return nil, errors.Errorf("unsupport expr function: %v", callExpr.Function)
	}
}

func flattenWorkspaceApprovalSetting(ctx context.Context, client api.Client, setting *v1pb.WorkspaceApprovalSetting) ([]interface{}, error) {
	ruleList := []interface{}{}
	for _, rule := range setting.Rules {
		stepList := []interface{}{}
		for _, step := range rule.Template.Flow.Steps {
			nodeList := []interface{}{}
			for _, node := range step.Nodes {
				rawNode := map[string]interface{}{
					"type":             node.Type.String(),
					"group_value":      node.GetGroupValue().String(),
					"role":             node.GetRole(),
					"external_node_id": node.GetExternalNodeId(),
				}
				nodeList = append(nodeList, rawNode)
			}
			raw := map[string]interface{}{
				"type":  step.Type.String(),
				"nodes": nodeList,
			}
			stepList = append(stepList, raw)
		}

		conditionList := []map[string]interface{}{}
		if rule.Condition.Expression != "" {
			tflog.Debug(ctx, rule.Condition.Expression)

			parsedExpr, err := client.ParseExpression(ctx, rule.Condition.Expression)
			if err != nil {
				return nil, err
			}
			expressions, err := parseApprovalExpression(parsedExpr.GetCallExpr())
			if err != nil {
				return nil, err
			}
			conditionList = expressions
		}

		raw := map[string]interface{}{
			"conditions": conditionList,
			"flow": []interface{}{
				map[string]interface{}{
					"title":       rule.Template.Title,
					"description": rule.Template.Description,
					"creator":     rule.Template.Creator,
					"steps":       stepList,
				},
			},
		}

		ruleList = append(ruleList, raw)
	}

	approvalSetting := map[string]interface{}{
		"rules": ruleList,
	}
	return []interface{}{approvalSetting}, nil
}
