package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
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
				Type:        schema.TypeString,
				Required:    true,
				Description: `The setting name in settings/{name} format. The name support "WORKSPACE_APPROVAL", "WORKSPACE_PROFILE", "DATA_CLASSIFICATION", "SEMANTIC_TYPES", "ENVIRONMENT", "PASSWORD_RESTRICTION", "SQL_RESULT_SIZE_LIMIT". Check the proto https://github.com/bytebase/bytebase/blob/main/proto/v1/v1/setting_service.proto#L109 for details`,
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s%s$", internal.SettingNamePrefix, v1pb.Setting_WORKSPACE_APPROVAL.String()),
					fmt.Sprintf("^%s%s$", internal.SettingNamePrefix, v1pb.Setting_WORKSPACE_PROFILE.String()),
					fmt.Sprintf("^%s%s$", internal.SettingNamePrefix, v1pb.Setting_DATA_CLASSIFICATION.String()),
					fmt.Sprintf("^%s%s$", internal.SettingNamePrefix, v1pb.Setting_SEMANTIC_TYPES.String()),
					fmt.Sprintf("^%s%s$", internal.SettingNamePrefix, v1pb.Setting_ENVIRONMENT.String()),
					fmt.Sprintf("^%s%s$", internal.SettingNamePrefix, v1pb.Setting_PASSWORD_RESTRICTION.String()),
				),
			},
			"approval_flow":        getWorkspaceApprovalSetting(true),
			"workspace_profile":    getWorkspaceProfileSetting(true),
			"classification":       getClassificationSetting(true),
			"semantic_types":       getSemanticTypesSetting(true),
			"environment_setting":  getEnvironmentSetting(true),
			"password_restriction": getPasswordRestrictionSetting(true),
		},
	}
}

func getSemanticTypesSetting(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed:    computed,
		Optional:    true,
		Default:     nil,
		Type:        schema.TypeSet,
		Description: "Semantic types for data masking. Require ENTERPRISE subscription.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The semantic type unique uuid.",
				},
				"title": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The semantic type title. Required.",
				},
				"description": {
					Type:        schema.TypeString,
					Computed:    computed,
					Optional:    true,
					Description: "The semantic type description. Optional.",
				},
				"algorithm": {
					Type:        schema.TypeList,
					Computed:    computed,
					Optional:    true,
					MaxItems:    1,
					MinItems:    0,
					Description: "The semantic type algorithm. Required.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"full_mask": {
								Type:     schema.TypeList,
								Computed: computed,
								Optional: true,
								MaxItems: 1,
								MinItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"substitution": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "Substitution is the string used to replace the original value, the max length of the string is 16 bytes.",
										},
									},
								},
							},
							"range_mask": {
								Type:     schema.TypeList,
								Computed: computed,
								Optional: true,
								MaxItems: 1,
								MinItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"slices": {
											Type:     schema.TypeList,
											Computed: computed,
											Optional: true,
											MinItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"start": {
														Type:        schema.TypeInt,
														Required:    true,
														Description: "Start is the start index of the original value, start from 0 and should be less than stop.",
													},
													"end": {
														Type:        schema.TypeInt,
														Required:    true,
														Description: "End is the stop index of the original value, should be less than the length of the original value.",
													},
													"substitution": {
														Type:        schema.TypeString,
														Required:    true,
														Description: "Substitution is the string used to replace the OriginalValue[start:end).",
													},
												},
											},
										},
									},
								},
							},
							"md5_mask": {
								Type:     schema.TypeList,
								Computed: computed,
								Optional: true,
								MaxItems: 1,
								MinItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"salt": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "Salt is the salt value to generate a different hash that with the word alone.",
										},
									},
								},
							},
							"inner_outer_mask": {
								Type:     schema.TypeList,
								Computed: computed,
								Optional: true,
								MaxItems: 1,
								MinItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"prefix_len": {
											Type:        schema.TypeInt,
											Required:    true,
											Description: "The length of prefix.",
										},
										"suffix_len": {
											Type:        schema.TypeInt,
											Required:    true,
											Description: "The length of suffix.",
										},
										"substitution": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "Substitution is the string used to replace the inner or outer substring.",
										},
										"type": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "INNER or OUTER.",
											ValidateFunc: validation.StringInSlice([]string{
												v1pb.Algorithm_InnerOuterMask_INNER.String(),
												v1pb.Algorithm_InnerOuterMask_OUTER.String(),
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
		Set: semanticTypeHash,
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
					Required:    true,
					Description: "The classification unique uuid.",
				},
				"title": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The classification title. Optional.",
				},
				"classification_from_config": {
					Type:        schema.TypeBool,
					Computed:    computed,
					Optional:    true,
					Description: "If true, we will only store the classification in the config. Otherwise we will get the classification from table/column comment, and write back to the schema metadata.",
				},
				"levels": {
					Required: true,
					Type:     schema.TypeSet,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The classification level unique uuid.",
							},
							"title": {
								Type:        schema.TypeString,
								Required:    true,
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
					Set: levelHash,
				},
				"classifications": {
					Required: true,
					Type:     schema.TypeSet,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The classification unique id, must in {number}-{number} format.",
							},
							"title": {
								Type:        schema.TypeString,
								Required:    true,
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
					Set: classificationHash,
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
					Optional:    true,
					Description: "The URL user visits Bytebase. The external URL is used for: 1. Constructing the correct callback URL when configuring the VCS provider. The callback URL points to the frontend; 2. Creating the correct webhook endpoint when configuring the project GitOps workflow. The webhook endpoint points to the backend.",
				},
				"disallow_signup": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Disallow self-service signup, users can only be invited by the owner. Require PRO subscription.",
				},
				"disallow_password_signin": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Whether to disallow password signin (except workspace admins). Require ENTERPRISE subscription",
				},
				"domains": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "The workspace domain, e.g. bytebase.com. Required for the group",
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"enforce_identity_domain": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Only user and group from the domains can be created and login.",
				},
				"database_change_mode": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  v1pb.DatabaseChangeMode_DATABASE_CHANGE_MODE_UNSPECIFIED.String(),
					ValidateFunc: validation.StringInSlice([]string{
						v1pb.DatabaseChangeMode_EDITOR.String(),
						v1pb.DatabaseChangeMode_PIPELINE.String(),
					}, false),
					Description: "The workspace database change mode, support EDITOR or PIPELINE. Default PIPELINE",
				},
				"token_duration_in_seconds": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: validation.IntAtLeast(3600),
					Description:  "The duration for login token in seconds. The duration should be at least 3600 (one hour).",
				},
				"maximum_role_expiration_in_seconds": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "The max duration in seconds for role expired. If the value is less than or equal to 0, we will remove the setting. AKA no limit.",
				},
				"announcement": {
					Type:        schema.TypeList,
					Optional:    true,
					MinItems:    0,
					MaxItems:    1,
					Description: "Custom announcement. Will show as a banner in the Bytebase UI. Require ENTERPRISE subscription.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"text": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The text of announcement. Leave it as empty string can clear the announcement",
							},
							"link": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The optional link, user can follow the link to check extra details",
							},
							"level": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The alert level of announcement",
								Default:     v1pb.Announcement_ALERT_LEVEL_UNSPECIFIED.String(),
								ValidateFunc: validation.StringInSlice([]string{
									v1pb.Announcement_INFO.String(),
									v1pb.Announcement_WARNING.String(),
									v1pb.Announcement_CRITICAL.String(),
								}, false),
							},
						},
					},
				},
			},
		},
	}
}

const minimumPasswordLength = 8

func getPasswordRestrictionSetting(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed:    computed,
		Optional:    true,
		Default:     nil,
		Type:        schema.TypeList,
		MaxItems:    1,
		MinItems:    1,
		Description: "Restrict for login password",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"min_length": {
					Type:         schema.TypeInt,
					Optional:     true,
					Description:  fmt.Sprintf("min_length is the minimum length for password, should no less than %d.", minimumPasswordLength),
					ValidateFunc: validation.IntAtLeast(minimumPasswordLength),
				},
				"require_number": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "require_number requires the password must contains at least one number.",
				},
				"require_letter": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "require_letter requires the password must contains at least one letter, regardless of upper case or lower case.",
				},
				"require_uppercase_letter": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "require_uppercase_letter requires the password must contains at least one upper case letter.",
				},
				"require_special_character": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "require_special_character requires the password must contains at least one special character.",
				},
				"require_reset_password_for_first_login": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "require_reset_password_for_first_login requires users to reset their password after the 1st login.",
				},
				"password_rotation_in_seconds": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: validation.IntAtLeast(86400),
					Description:  "password_rotation requires users to reset their password after the duration. The duration should be at least 86400 (one day).",
				},
			},
		},
	}
}

func getEnvironmentSetting(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed:    computed,
		Optional:    true,
		Default:     nil,
		Type:        schema.TypeList,
		Description: "The environment",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"environment": {
					Type:     schema.TypeList,
					Computed: computed,
					Required: !computed,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: internal.ResourceIDValidation,
								Description:  "The environment unique id.",
							},
							"name": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The environment readonly name in environments/{id} format.",
							},
							"title": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringIsNotEmpty,
								Description:  "The environment display name.",
							},
							"color": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The environment color.",
							},
							"protected": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "The environment is protected or not.",
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
										"id": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringIsNotEmpty,
											Description:  "The approval template ID. Built-in templates use 'bb.*' prefix (e.g., 'bb.project-owner', 'bb.workspace-dba'), custom templates use UUIDs.",
										},
										"title": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringIsNotEmpty,
										},
										"description": {
											Type:     schema.TypeString,
											Computed: computed,
											Optional: true,
										},
										"roles": {
											Type:        schema.TypeList,
											Required:    true,
											Description: "The role require to review in this step",
											Elem: &schema.Schema{
												Type:        schema.TypeString,
												Description: `Role full name in roles/{id} format.`,
												ValidateDiagFunc: internal.ResourceNameValidation(
													fmt.Sprintf("^%s", internal.RoleNamePrefix),
												),
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
												v1pb.Risk_REQUEST_ROLE.String(),
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

	settingName := d.Get("name").(string)
	setting, err := c.GetSetting(ctx, settingName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(setting.Name)
	return setSettingMessage(ctx, d, c, setting)
}

func setSettingMessage(ctx context.Context, d *schema.ResourceData, client api.Client, setting *v1pb.Setting) diag.Diagnostics {
	if value := setting.GetValue().GetWorkspaceApprovalSettingValue(); value != nil {
		settingVal, err := flattenWorkspaceApprovalSetting(ctx, client, value)
		if err != nil {
			return diag.Errorf("failed to parse workspace_approval_setting: %s", err.Error())
		}
		if err := d.Set("approval_flow", settingVal); err != nil {
			return diag.Errorf("cannot set workspace_approval_setting: %s", err.Error())
		}
	}
	if value := setting.GetValue().GetWorkspaceProfileSettingValue(); value != nil {
		settingVal := flattenWorkspaceProfileSetting(value)
		if err := d.Set("workspace_profile", settingVal); err != nil {
			return diag.Errorf("cannot set workspace_profile: %s", err.Error())
		}
	}
	if value := setting.GetValue().GetPasswordRestrictionSetting(); value != nil {
		settingVal := flattenPasswordRestrictionSetting(value)
		if err := d.Set("password_restriction", settingVal); err != nil {
			return diag.Errorf("cannot set password_restriction: %s", err.Error())
		}
	}
	if value := setting.GetValue().GetDataClassificationSettingValue(); value != nil {
		settingVal := flattenClassificationSetting(value)
		if err := d.Set("classification", settingVal); err != nil {
			return diag.Errorf("cannot set classification: %s", err.Error())
		}
	}
	if value := setting.GetValue().GetSemanticTypeSettingValue(); value != nil {
		settingVal := flattenSemanticTypesSetting(value)
		if err := d.Set("semantic_types", schema.NewSet(semanticTypeHash, settingVal)); err != nil {
			return diag.Errorf("cannot set semantic_types: %s", err.Error())
		}
	}
	if value := setting.GetValue().GetEnvironmentSetting(); value != nil {
		settingVal := flattenEnvironmentSetting(value)
		if err := d.Set("environment_setting", settingVal); err != nil {
			return diag.Errorf("cannot set environment_setting: %s", err.Error())
		}
	}

	return nil
}

func flattenEnvironmentSetting(setting *v1pb.EnvironmentSetting) []interface{} {
	environmentList := []interface{}{}

	for _, environment := range setting.GetEnvironments() {
		raw := map[string]interface{}{
			"id":        environment.Id,
			"name":      environment.Name,
			"color":     environment.Color,
			"title":     environment.Title,
			"protected": environment.Tags["protected"] == "protected",
		}
		environmentList = append(environmentList, raw)
	}

	environmentSetting := map[string]interface{}{
		"environment": environmentList,
	}
	return []interface{}{environmentSetting}
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
		roleList := []interface{}{}
		for _, role := range rule.Template.Flow.Roles {
			roleList = append(roleList, role)
		}

		conditionList := []map[string]interface{}{}
		if rule.Condition.Expression != "" {
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
					"id":          rule.Template.Id,
					"title":       rule.Template.Title,
					"description": rule.Template.Description,
					"roles":       roleList,
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

func flattenWorkspaceProfileSetting(setting *v1pb.WorkspaceProfileSetting) []interface{} {
	raw := map[string]interface{}{}

	raw["external_url"] = setting.ExternalUrl
	raw["disallow_signup"] = setting.DisallowSignup
	raw["disallow_password_signin"] = setting.DisallowPasswordSignin
	raw["enforce_identity_domain"] = setting.EnforceIdentityDomain
	raw["domains"] = setting.Domains
	raw["database_change_mode"] = setting.DatabaseChangeMode.String()

	if v := setting.GetTokenDuration(); v != nil {
		raw["token_duration_in_seconds"] = int(v.Seconds)
	}
	if v := setting.GetMaximumRoleExpiration(); v != nil {
		raw["maximum_role_expiration_in_seconds"] = int(v.Seconds)
	}
	// Handle announcement field - need to be careful with empty announcements
	if v := setting.GetAnnouncement(); v != nil {
		// Check if this is truly an empty announcement (all fields at their zero/default values)
		isEmpty := v.Text == "" && v.Link == "" && v.Level == v1pb.Announcement_ALERT_LEVEL_UNSPECIFIED
		if !isEmpty {
			raw["announcement"] = []any{
				map[string]any{
					"text":  v.Text,
					"link":  v.Link,
					"level": v.Level.String(),
				},
			}
		}
		// If announcement is empty, don't set it at all - let Terraform handle it as unset
	}

	return []interface{}{raw}
}

func flattenPasswordRestrictionSetting(setting *v1pb.PasswordRestrictionSetting) []interface{} {
	raw := map[string]interface{}{}
	raw["min_length"] = int(setting.MinLength)
	raw["require_number"] = setting.RequireNumber
	raw["require_letter"] = setting.RequireLetter
	raw["require_uppercase_letter"] = setting.RequireUppercaseLetter
	raw["require_special_character"] = setting.RequireSpecialCharacter
	raw["require_reset_password_for_first_login"] = setting.RequireResetPasswordForFirstLogin
	if v := setting.GetPasswordRotation(); v != nil {
		raw["password_rotation_in_seconds"] = int(v.Seconds)
	}
	return []interface{}{raw}
}

func flattenClassificationSetting(setting *v1pb.DataClassificationSetting) []interface{} {
	raw := map[string]interface{}{}

	if len(setting.GetConfigs()) > 0 {
		config := setting.GetConfigs()[0]
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
		raw["levels"] = schema.NewSet(levelHash, rawLevels)

		rawClassifications := []interface{}{}
		for _, classification := range config.GetClassification() {
			rawClassification := map[string]interface{}{}
			rawClassification["id"] = classification.Id
			rawClassification["title"] = classification.Title
			rawClassification["description"] = classification.Description
			rawClassification["level"] = classification.LevelId
			rawClassifications = append(rawClassifications, rawClassification)
		}
		raw["classifications"] = schema.NewSet(classificationHash, rawClassifications)
	}

	return []interface{}{raw}
}

func flattenSemanticTypesSetting(setting *v1pb.SemanticTypeSetting) []interface{} {
	raw := []interface{}{}

	for _, semanticType := range setting.Types {
		rawData := map[string]interface{}{}
		rawData["id"] = semanticType.Id
		rawData["title"] = semanticType.Title
		rawData["description"] = semanticType.Description

		if v := semanticType.Algorithm; v != nil {
			switch m := v.Mask.(type) {
			case *v1pb.Algorithm_FullMask_:
				fullMask := map[string]interface{}{
					"substitution": m.FullMask.Substitution,
				}
				rawData["algorithm"] = []interface{}{
					map[string]interface{}{
						"full_mask": []interface{}{
							fullMask,
						},
					},
				}
			case *v1pb.Algorithm_RangeMask_:
				rangeMaskSlices := []interface{}{}
				for _, s := range m.RangeMask.Slices {
					rangeMaskSlice := map[string]interface{}{}
					rangeMaskSlice["start"] = s.Start
					rangeMaskSlice["end"] = s.End
					rangeMaskSlice["substitution"] = s.Substitution
					rangeMaskSlices = append(rangeMaskSlices, rangeMaskSlice)
				}
				rangeMask := map[string]interface{}{
					"slices": rangeMaskSlices,
				}
				rawData["algorithm"] = []interface{}{
					map[string]interface{}{
						"range_mask": []interface{}{
							rangeMask,
						},
					},
				}
			case *v1pb.Algorithm_Md5Mask:
				md5Mask := map[string]interface{}{
					"salt": m.Md5Mask.Salt,
				}
				rawData["algorithm"] = []interface{}{
					map[string]interface{}{
						"md5_mask": []interface{}{
							md5Mask,
						},
					},
				}
			case *v1pb.Algorithm_InnerOuterMask_:
				innerOuterMask := map[string]interface{}{
					"prefix_len":   m.InnerOuterMask.PrefixLen,
					"suffix_len":   m.InnerOuterMask.SuffixLen,
					"substitution": m.InnerOuterMask.Substitution,
					"type":         m.InnerOuterMask.Type.String(),
				}
				rawData["algorithm"] = []interface{}{
					map[string]interface{}{
						"inner_outer_mask": []interface{}{
							innerOuterMask,
						},
					},
				}
			}
		}

		raw = append(raw, rawData)
	}

	return raw
}

func levelHash(rawSchema interface{}) int {
	classificationLevel := convertToV1Level(rawSchema)
	return internal.ToHash(classificationLevel)
}

func classificationHash(rawSchema interface{}) int {
	classificationData := convertToV1Classification(rawSchema)
	return internal.ToHash(classificationData)
}

func semanticTypeHash(rawSchema interface{}) int {
	semanticType, err := convertToV1SemanticType(rawSchema)
	if err != nil {
		return 0
	}
	return internal.ToHash(semanticType)
}
