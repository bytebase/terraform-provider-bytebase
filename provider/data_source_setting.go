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
					string(api.SettingWorkspaceProfile),
					string(api.SettingDataClassification),
				}, false),
			},
			"approval_flow":           getWorkspaceApprovalSetting(true),
			"external_approval_nodes": getExternalApprovalSetting(true),
			"workspace_profile":       getWorkspaceProfileSetting(true),
			"classification":          getClassificationSetting(true),
		},
	}
}

func getClassificationSetting(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed:    computed,
		Optional:    true,
		Default:     nil,
		Type:        schema.TypeList,
		MaxItems:    1,
		MinItems:    1,
		Description: "Classification for data masking. Require ENTERPRISE subscription.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    computed,
					Optional:    true,
					Description: "The classification unique uuid.",
				},
				"title": {
					Type:        schema.TypeString,
					Computed:    computed,
					Optional:    true,
					Description: "The classification title. Optional.",
				},
				"classification_from_config": {
					Type:        schema.TypeBool,
					Computed:    computed,
					Optional:    true,
					Description: "If true, we will only store the classification in the config. Otherwise we will get the classification from table/column comment, and write back to the schema metadata.",
				},
				"levels": {
					Computed: computed,
					Optional: true,
					Type:     schema.TypeList,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:        schema.TypeString,
								Computed:    computed,
								Optional:    true,
								Description: "The classification level unique uuid.",
							},
							"title": {
								Type:        schema.TypeString,
								Computed:    computed,
								Optional:    true,
								Description: "The classification level title.",
							},
							"description": {
								Type:        schema.TypeString,
								Computed:    computed,
								Optional:    true,
								Description: "The classification level description.",
							},
						},
					},
				},
				"classifications": {
					Computed: computed,
					Optional: true,
					Type:     schema.TypeList,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:        schema.TypeString,
								Computed:    computed,
								Optional:    true,
								Description: "The classification unique id, must in {number}-{number} format.",
							},
							"title": {
								Type:        schema.TypeString,
								Computed:    computed,
								Optional:    true,
								Description: "The classification title.",
							},
							"description": {
								Type:        schema.TypeString,
								Computed:    computed,
								Optional:    true,
								Description: "The classification description.",
							},
							"level": {
								Type:        schema.TypeString,
								Computed:    computed,
								Optional:    true,
								Description: "The classification level id.",
							},
						},
					},
				},
			},
		},
	}
}

func getWorkspaceProfileSetting(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed: computed,
		Optional: true,
		Default:  nil,
		Type:     schema.TypeList,
		MaxItems: 1,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"external_url": {
					Type:        schema.TypeString,
					Computed:    computed,
					Optional:    true,
					Description: "The URL user visits Bytebase. The external URL is used for: 1. Constructing the correct callback URL when configuring the VCS provider. The callback URL points to the frontend; 2. Creating the correct webhook endpoint when configuring the project GitOps workflow. The webhook endpoint points to the backend.",
				},
				"disallow_signup": {
					Type:        schema.TypeBool,
					Computed:    computed,
					Optional:    true,
					Description: "Disallow self-service signup, users can only be invited by the owner. Require PRO subscription.",
				},
				"disallow_password_signin": {
					Type:        schema.TypeBool,
					Computed:    computed,
					Optional:    true,
					Description: "Whether to disallow password signin. (Except workspace admins). Require ENTERPRISE subscription",
				},
				"domains": {
					Type:        schema.TypeList,
					Computed:    computed,
					Optional:    true,
					Description: "The workspace domain, e.g. bytebase.com. Required for the group",
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"enforce_identity_domain": {
					Type:        schema.TypeBool,
					Computed:    computed,
					Optional:    true,
					Description: "Only user and group from the domains can be created and login.",
				},
			},
		},
	}
}

func getExternalApprovalSetting(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed:    computed,
		Optional:    true,
		Default:     nil,
		Type:        schema.TypeList,
		Description: "Configure external nodes in the approval flow. Require ENTERPRISE subscription.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"nodes": {
					Type:     schema.TypeList,
					Computed: computed,
					Required: !computed,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:        schema.TypeString,
								Computed:    computed,
								Required:    !computed,
								Description: "The unique external node id.",
							},
							"title": {
								Type:        schema.TypeString,
								Computed:    computed,
								Required:    !computed,
								Description: "The external node title.",
							},
							"endpoint": {
								Type:        schema.TypeString,
								Computed:    computed,
								Required:    !computed,
								Description: "The endpoint URL to receive the approval message. Learn more: https://www.bytebase.com/docs/api/external-approval",
							},
						},
					},
				},
			},
		},
	}
}

func getWorkspaceApprovalSetting(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed:    computed,
		Optional:    true,
		Default:     nil,
		Type:        schema.TypeList,
		Description: "Configure risk level and approval flow for different tasks. Require ENTERPRISE subscription.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"rules": {
					Type:     schema.TypeList,
					Computed: computed,
					Required: !computed,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"flow": {
								Computed: computed,
								Required: !computed,
								Type:     schema.TypeList,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"title": {
											Type:     schema.TypeString,
											Computed: computed,
											Required: !computed,
										},
										"description": {
											Type:     schema.TypeString,
											Computed: computed,
											Optional: true,
										},
										"creator": {
											Type:        schema.TypeString,
											Computed:    computed,
											Required:    !computed,
											Description: "The creator name in users/{email} format",
										},
										"steps": {
											Type:        schema.TypeList,
											Computed:    computed,
											Required:    !computed,
											Description: "Approval flow following the step order.",
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"type": {
														Type:     schema.TypeString,
														Computed: computed,
														Optional: true,
														ValidateFunc: validation.StringInSlice([]string{
															string(api.ApprovalNodeTypeGroup),
															string(api.ApprovalNodeTypeRole),
															string(api.ApprovalNodeTypeExternalNodeID),
														}, false),
													},
													"node": {
														Required: !computed,
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
							"conditions": {
								MinItems:    0,
								Computed:    computed,
								Type:        schema.TypeList,
								Optional:    true,
								Description: "Match any condition will trigger this approval flow.",
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
		if err := d.Set("approval_flow", settingVal); err != nil {
			return diag.Errorf("cannot set workspace_approval_setting: %s", err.Error())
		}
	}
	if value := setting.Value.GetExternalApprovalSettingValue(); value != nil {
		settingVal := flattenExternalApprovalSetting(value)
		if err := d.Set("external_approval_nodes", settingVal); err != nil {
			return diag.Errorf("cannot set external_approval_nodes: %s", err.Error())
		}
	}
	if value := setting.Value.GetWorkspaceProfileSettingValue(); value != nil {
		settingVal := flattenWorkspaceProfileSetting(value)
		if err := d.Set("workspace_profile", settingVal); err != nil {
			return diag.Errorf("cannot set workspace_profile: %s", err.Error())
		}
	}
	if value := setting.Value.GetDataClassificationSettingValue(); value != nil {
		settingVal := flattenClassificationSetting(value)
		if err := d.Set("classification", settingVal); err != nil {
			return diag.Errorf("cannot set classification: %s", err.Error())
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
				switch int(levelNumber) {
				case api.RiskLevelDefault.Int():
					resp[argName] = api.RiskLevelDefault
				case api.RiskLevelLow.Int():
					resp[argName] = api.RiskLevelLow
				case api.RiskLevelModerate.Int():
					resp[argName] = api.RiskLevelModerate
				case api.RiskLevelHigh.Int():
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
			for _, node := range step.Nodes {
				rawNode := map[string]interface{}{}
				switch payload := node.Payload.(type) {
				case *v1pb.ApprovalNode_Role:
					rawNode["type"] = string(api.ApprovalNodeTypeRole)
					rawNode["node"] = payload.Role
				case *v1pb.ApprovalNode_ExternalNodeId:
					rawNode["type"] = string(api.ApprovalNodeTypeExternalNodeID)
					rawNode["node"] = payload.ExternalNodeId
				case *v1pb.ApprovalNode_GroupValue_:
					rawNode["type"] = string(api.ApprovalNodeTypeGroup)
					rawNode["node"] = payload.GroupValue.String()
				}
				stepList = append(stepList, rawNode)
			}
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

func flattenExternalApprovalSetting(setting *v1pb.ExternalApprovalSetting) []interface{} {
	nodeList := []interface{}{}
	for _, node := range setting.Nodes {
		rawNode := map[string]interface{}{}
		rawNode["id"] = node.Id
		rawNode["title"] = node.Title
		rawNode["endpoint"] = node.Endpoint
		nodeList = append(nodeList, rawNode)
	}

	approvalSetting := map[string]interface{}{
		"nodes": nodeList,
	}
	return []interface{}{approvalSetting}
}

func flattenWorkspaceProfileSetting(setting *v1pb.WorkspaceProfileSetting) []interface{} {
	raw := map[string]interface{}{}

	raw["external_url"] = setting.ExternalUrl
	raw["disallow_signup"] = setting.DisallowSignup
	raw["disallow_password_signin"] = setting.DisallowPasswordSignin
	raw["enforce_identity_domain"] = setting.EnforceIdentityDomain
	raw["domains"] = setting.Domains

	return []interface{}{raw}
}

func flattenClassificationSetting(setting *v1pb.DataClassificationSetting) []interface{} {
	raw := map[string]interface{}{}

	for _, config := range setting.GetConfigs() {
		raw["id"] = config.Id
		raw["title"] = config.Title
		raw["classification_from_config"] = config.ClassificationFromConfig

		rawLevels := []interface{}{}
		for _, level := range config.Levels {
			rawLevel := map[string]interface{}{}
			rawLevel["id"] = level.Id
			rawLevel["title"] = level.Title
			rawLevel["description"] = level.Description
			rawLevels = append(rawLevels, rawLevel)
		}
		raw["levels"] = rawLevels

		rawClassifications := []interface{}{}
		for _, classification := range config.GetClassification() {
			rawClassification := map[string]interface{}{}
			rawClassification["id"] = classification.Id
			rawClassification["title"] = classification.Title
			rawClassification["description"] = classification.Description
			rawClassification["level"] = classification.LevelId
			rawClassifications = append(rawClassifications, rawClassification)
		}
		raw["classifications"] = rawClassifications
		break
	}

	return []interface{}{raw}
}
