package provider

import (
	"context"
	"fmt"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/genproto/googleapis/type/expr"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceDatabaseGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "The database group resource.",
		ReadContext:   resourceDatabaseGroupRead,
		DeleteContext: resourceDatabaseGroupDelete,
		CreateContext: resourceDatabaseGroupCreate,
		UpdateContext: resourceDatabaseGroupUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The database group unique resource id.",
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern),
				),
				Description: "The project fullname in projects/{id} format.",
			},
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The database group title.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"condition": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The database group condition. Check the proto message https://github.com/bytebase/bytebase/blob/main/proto/v1/v1/database_group_service.proto#L185 for details.",
			},
			"matched_databases": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The matched databases in the group.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceDatabaseGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	fullName := d.Id()

	group, err := c.GetDatabaseGroup(ctx, fullName, v1pb.DatabaseGroupView_DATABASE_GROUP_VIEW_FULL)
	if err != nil {
		// Check if the resource was deleted outside of Terraform
		if internal.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Resource %s not found, removing from state", fullName))
			// Remove from state to trigger recreation on next apply
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return setDatabaseGroup(d, group)
}

func setDatabaseGroup(d *schema.ResourceData, group *v1pb.DatabaseGroup) diag.Diagnostics {
	projectID, groupID, err := internal.GetProjectDatabaseGroupID(group.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("resource_id", groupID); err != nil {
		return diag.Errorf("cannot set resource_id for group: %s", err.Error())
	}
	if err := d.Set("project", fmt.Sprintf("%s%s", internal.ProjectNamePrefix, projectID)); err != nil {
		return diag.Errorf("cannot set resource_id for group: %s", err.Error())
	}
	if err := d.Set("title", group.Title); err != nil {
		return diag.Errorf("cannot set title for group: %s", err.Error())
	}
	if err := d.Set("condition", group.DatabaseExpr.Expression); err != nil {
		return diag.Errorf("cannot set condition for group: %s", err.Error())
	}

	matchedDatabases := []string{}
	for _, db := range group.MatchedDatabases {
		matchedDatabases = append(matchedDatabases, db.Name)
	}
	if err := d.Set("matched_databases", matchedDatabases); err != nil {
		return diag.Errorf("cannot set matched_databases for group: %s", err.Error())
	}

	return nil
}

func resourceDatabaseGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	groupID := d.Get("resource_id").(string)
	projectName := d.Get("project").(string)
	groupName := fmt.Sprintf("%s/%s%s", projectName, internal.DatabaseGroupNamePrefix, groupID)

	existedGroup, err := c.GetDatabaseGroup(ctx, groupName, v1pb.DatabaseGroupView_DATABASE_GROUP_VIEW_BASIC)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get group %s failed with error: %v", groupName, err))
	}

	databaseGroup := &v1pb.DatabaseGroup{
		Name:  groupName,
		Title: d.Get("title").(string),
		DatabaseExpr: &expr.Expr{
			Expression: d.Get("condition").(string),
		},
	}

	var diags diag.Diagnostics
	if existedGroup != nil && err == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Database group already exists",
			Detail:   fmt.Sprintf("Database group %s already exists, try to exec the update operation", groupName),
		})

		updated, err := c.UpdateDatabaseGroup(ctx, databaseGroup, []string{"title", "database_expr"})
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update database group",
				Detail:   fmt.Sprintf("Update database group %s failed, error: %v", groupName, err),
			})
			return diags
		}
		existedGroup = updated
	} else {
		created, err := c.CreateDatabaseGroup(ctx, projectName, groupID, databaseGroup)
		if err != nil {
			return diag.FromErr(err)
		}
		existedGroup = created
	}

	d.SetId(existedGroup.Name)

	diag := setDatabaseGroup(d, existedGroup)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceDatabaseGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("resource_id") {
		return diag.Errorf("cannot change the resource id")
	}
	if d.HasChange("project") {
		return diag.Errorf("cannot change the project")
	}

	c := m.(api.Client)
	groupName := d.Id()

	databaseGroup := &v1pb.DatabaseGroup{
		Name:  groupName,
		Title: d.Get("title").(string),
		DatabaseExpr: &expr.Expr{
			Expression: d.Get("condition").(string),
		},
	}

	updateMasks := []string{}
	if d.HasChange("title") {
		updateMasks = append(updateMasks, "title")
	}
	if d.HasChange("condition") {
		updateMasks = append(updateMasks, "database_expr")
	}

	if len(updateMasks) == 0 {
		return nil
	}

	var diags diag.Diagnostics
	updated, err := c.UpdateDatabaseGroup(ctx, databaseGroup, updateMasks)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to update database group",
			Detail:   fmt.Sprintf("Update database group %s failed, error: %v", groupName, err),
		})
		return diags
	}
	databaseGroup = updated

	diag := setDatabaseGroup(d, databaseGroup)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceDatabaseGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	return internal.ResourceDelete(ctx, d, c.DeleteDatabaseGroup)
}
