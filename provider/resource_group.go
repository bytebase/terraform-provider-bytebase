package provider

import (
	"context"
	"fmt"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "The group resource. Workspace domain is required for creating groups.",
		ReadContext:   internal.ResourceRead(resourceGroupRead),
		DeleteContext: internal.ResourceDelete,
		CreateContext: resourceGroupCreate,
		UpdateContext: resourceGroupUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"email": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The group email.",
			},
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The group title.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The group description.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The group name in groups/{email} format.",
			},
			"source": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Source means where the group comes from. For now we support Entra ID SCIM sync, so the source could be Entra ID.",
			},
			"members": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Description: "The members in the group.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"member": {
							Type:     schema.TypeString,
							Required: true,
							ValidateDiagFunc: internal.ResourceNameValidation(
								fmt.Sprintf("^%s", internal.UserNamePrefix),
							),
							Description: "The member in users/{email} format.",
						},
						"role": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The member's role in the group, should be OWNER or MEMBER.",
							ValidateFunc: validation.StringInSlice([]string{
								v1pb.GroupMember_OWNER.String(),
								v1pb.GroupMember_MEMBER.String(),
							}, false),
						},
					},
				},
				Set: memberHash,
			},
		},
	}
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	fullName := d.Id()
	group, err := c.GetGroup(ctx, fullName)
	if err != nil {
		return diag.FromErr(err)
	}

	return setGroup(d, group)
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	groupEmail := d.Get("email").(string)
	groupName := fmt.Sprintf("%s%s", internal.GroupNamePrefix, groupEmail)

	members, err := convertToMemberList(d)
	if err != nil {
		return diag.FromErr(err)
	}

	group := &v1pb.Group{
		Name:        groupName,
		Title:       d.Get("title").(string),
		Description: d.Get("description").(string),
		Members:     members,
	}

	existedGroup, err := c.GetGroup(ctx, groupName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get group %s failed with error: %v", groupName, err))
	}

	var diags diag.Diagnostics
	if existedGroup != nil && err == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Group already exists",
			Detail:   fmt.Sprintf("Group %s already exists, try to exec the update operation", groupName),
		})

		updateMasks := []string{"members"}
		if group.Title != existedGroup.Title {
			updateMasks = append(updateMasks, "title")
		}
		rawConfig := d.GetRawConfig()
		if config := rawConfig.GetAttr("description"); !config.IsNull() && group.Description != existedGroup.Description {
			updateMasks = append(updateMasks, "description")
		}

		if _, err := c.UpdateGroup(ctx, group, updateMasks); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update group",
				Detail:   fmt.Sprintf("Update group %s failed, error: %v", groupName, err),
			})
			return diags
		}
	} else {
		if _, err := c.CreateGroup(ctx, groupEmail, group); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(groupName)

	diag := resourceGroupRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("email") {
		return diag.Errorf("cannot change the group email")
	}

	c := m.(api.Client)
	groupName := d.Id()

	title := d.Get("title").(string)
	description := d.Get("description").(string)
	members, err := convertToMemberList(d)
	if err != nil {
		return diag.FromErr(err)
	}

	updateMasks := []string{}
	if d.HasChange("title") {
		updateMasks = append(updateMasks, "title")
	}
	if d.HasChange("description") {
		updateMasks = append(updateMasks, "description")
	}
	if d.HasChange("members") {
		updateMasks = append(updateMasks, "members")
	}

	var diags diag.Diagnostics
	if len(updateMasks) > 0 {
		if _, err := c.UpdateGroup(ctx, &v1pb.Group{
			Name:        groupName,
			Title:       title,
			Description: description,
			Members:     members,
		}, updateMasks); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update group",
				Detail:   fmt.Sprintf("Update group %s failed, error: %v", groupName, err),
			})
			return diags
		}
	}

	diag := resourceGroupRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func convertToMemberList(d *schema.ResourceData) ([]*v1pb.GroupMember, error) {
	memberSet, ok := d.Get("members").(*schema.Set)
	if !ok {
		return nil, errors.Errorf("group members is required")
	}

	memberList := []*v1pb.GroupMember{}
	existOwner := false
	for _, m := range memberSet.List() {
		rawMember := m.(map[string]interface{})

		member := rawMember["member"].(string)
		role := v1pb.GroupMember_Role(v1pb.GroupMember_Role_value[rawMember["role"].(string)])
		memberList = append(memberList, &v1pb.GroupMember{
			Member: member,
			Role:   role,
		})

		if role == v1pb.GroupMember_OWNER {
			existOwner = true
		}
	}

	if !existOwner {
		return nil, errors.Errorf("require at least 1 group owner")
	}

	return memberList, nil
}
