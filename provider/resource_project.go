package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/type/expr"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
)

// defaultProj is the default project name.
var defaultProj = fmt.Sprintf("%sdefault", internal.ProjectNamePrefix)

func resourceProjct() *schema.Resource {
	return &schema.Resource{
		Description:          "The project resource.",
		CreateWithoutTimeout: resourceProjectCreate,
		ReadWithoutTimeout:   resourceProjectRead,
		UpdateWithoutTimeout: resourceProjectUpdate,
		DeleteContext:        resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The project unique resource id. Cannot change this after created.",
			},
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The project title.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project full name in projects/{resource id} format.",
			},
			"allow_modify_statement": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Allow modifying statement after issue is created.",
			},
			"auto_resolve_issue": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable auto resolve issue.",
			},
			"enforce_issue_title": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enforce issue title created by user instead of generated by Bytebase.",
			},
			"auto_enable_backup": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether to automatically enable backup.",
			},
			"skip_backup_errors": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether to skip backup errors and continue the data migration.",
			},
			"postgres_database_tenant_mode": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether to enable the database tenant mode for PostgreSQL. If enabled, the issue will be created with the pre-appended \"set role <db_owner>\" statement.",
			},
			"members":   getProjectMembersSchema(false),
			"databases": getDatabasesSchema(false),
		},
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	projectID := d.Get("resource_id").(string)
	projectName := fmt.Sprintf("%s%s", internal.ProjectNamePrefix, projectID)

	title := d.Get("title").(string)
	allowModifyStatement := d.Get("allow_modify_statement").(bool)
	autoResolveIssue := d.Get("auto_resolve_issue").(bool)
	enforceIssueTitle := d.Get("enforce_issue_title").(bool)
	autoEnableBackup := d.Get("auto_enable_backup").(bool)
	skipBackupErrors := d.Get("skip_backup_errors").(bool)
	postgresDatabaseTenantMode := d.Get("postgres_database_tenant_mode").(bool)

	existedProject, err := c.GetProject(ctx, projectName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get project %s failed with error: %v", projectName, err))
	}

	var diags diag.Diagnostics
	if existedProject != nil && err == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Project already exists",
			Detail:   fmt.Sprintf("Project %s already exists, try to exec the update operation", projectName),
		})

		if existedProject.State == v1pb.State_DELETED {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Project is deleted",
				Detail:   fmt.Sprintf("Project %s already deleted, try to undelete the project", projectName),
			})
			if _, err := c.UndeleteProject(ctx, projectName); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to undelete project",
					Detail:   fmt.Sprintf("Undelete project %s failed, error: %v", projectName, err),
				})
				return diags
			}
		}

		updateMasks := []string{}
		if title != "" && title != existedProject.Title {
			updateMasks = append(updateMasks, "title")
		}
		if allowModifyStatement != existedProject.AllowModifyStatement {
			updateMasks = append(updateMasks, "allow_modify_statement")
		}
		if autoResolveIssue != existedProject.AutoResolveIssue {
			updateMasks = append(updateMasks, "auto_resolve_issue")
		}
		if enforceIssueTitle != existedProject.EnforceIssueTitle {
			updateMasks = append(updateMasks, "enforce_issue_title")
		}
		if autoEnableBackup != existedProject.AutoEnableBackup {
			updateMasks = append(updateMasks, "auto_enable_backup")
		}
		if skipBackupErrors != existedProject.SkipBackupErrors {
			updateMasks = append(updateMasks, "skip_backup_errors")
		}
		if postgresDatabaseTenantMode != existedProject.PostgresDatabaseTenantMode {
			updateMasks = append(updateMasks, "postgres_database_tenant_mode")
		}

		if len(updateMasks) > 0 {
			if _, err := c.UpdateProject(ctx, &v1pb.Project{
				Name:                       projectName,
				Title:                      title,
				State:                      v1pb.State_ACTIVE,
				AllowModifyStatement:       allowModifyStatement,
				AutoResolveIssue:           autoResolveIssue,
				EnforceIssueTitle:          enforceIssueTitle,
				AutoEnableBackup:           autoEnableBackup,
				SkipBackupErrors:           skipBackupErrors,
				PostgresDatabaseTenantMode: postgresDatabaseTenantMode,
			}, updateMasks); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to update project",
					Detail:   fmt.Sprintf("Update project %s failed, error: %v", projectName, err),
				})
				return diags
			}
		}
	} else {
		if _, err := c.CreateProject(ctx, projectID, &v1pb.Project{
			Name:                       projectName,
			Title:                      title,
			State:                      v1pb.State_ACTIVE,
			AllowModifyStatement:       allowModifyStatement,
			AutoResolveIssue:           autoResolveIssue,
			EnforceIssueTitle:          enforceIssueTitle,
			AutoEnableBackup:           autoEnableBackup,
			SkipBackupErrors:           skipBackupErrors,
			PostgresDatabaseTenantMode: postgresDatabaseTenantMode,
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(projectName)

	if diag := updateDatabasesInProject(ctx, d, c, d.Id()); diag != nil {
		diags = append(diags, diag...)
		return diags
	}

	if diag := updateMembersInProject(ctx, d, c, d.Id()); diag != nil {
		diags = append(diags, diag...)
		return diags
	}

	tflog.Debug(ctx, "[upsert project] start reading project", map[string]interface{}{
		"project": projectName,
	})

	diag := resourceProjectRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	tflog.Debug(ctx, "[upsert project] upsert project finished", map[string]interface{}{
		"project": projectName,
	})

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("resource_id") {
		return diag.Errorf("cannot change the resource id")
	}

	c := m.(api.Client)
	projectName := d.Id()

	existedProject, err := c.GetProject(ctx, projectName)
	if err != nil {
		return diag.Errorf("get project %s failed with error: %v", projectName, err)
	}

	var diags diag.Diagnostics
	if existedProject.State == v1pb.State_DELETED {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Project is deleted",
			Detail:   fmt.Sprintf("Project %s already deleted, try to undelete the project", projectName),
		})
		if _, err := c.UndeleteProject(ctx, projectName); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to undelete project",
				Detail:   fmt.Sprintf("Undelete project %s failed, error: %v", projectName, err),
			})
			return diags
		}
	}

	paths := []string{}
	if d.HasChange("title") {
		paths = append(paths, "title")
	}
	if d.HasChange("allow_modify_statement") {
		paths = append(paths, "allow_modify_statement")
	}
	if d.HasChange("auto_resolve_issue") {
		paths = append(paths, "auto_resolve_issue")
	}
	if d.HasChange("enforce_issue_title") {
		paths = append(paths, "enforce_issue_title")
	}
	if d.HasChange("auto_enable_backup") {
		paths = append(paths, "auto_enable_backup")
	}
	if d.HasChange("skip_backup_errors") {
		paths = append(paths, "skip_backup_errors")
	}
	if d.HasChange("postgres_database_tenant_mode") {
		paths = append(paths, "postgres_database_tenant_mode")
	}

	allowModifyStatement := d.Get("allow_modify_statement").(bool)
	autoResolveIssue := d.Get("auto_resolve_issue").(bool)
	enforceIssueTitle := d.Get("enforce_issue_title").(bool)
	autoEnableBackup := d.Get("auto_enable_backup").(bool)
	skipBackupErrors := d.Get("skip_backup_errors").(bool)
	postgresDatabaseTenantMode := d.Get("postgres_database_tenant_mode").(bool)

	if len(paths) > 0 {
		if _, err := c.UpdateProject(ctx, &v1pb.Project{
			Name:                       projectName,
			Title:                      d.Get("title").(string),
			State:                      v1pb.State_ACTIVE,
			AllowModifyStatement:       allowModifyStatement,
			AutoResolveIssue:           autoResolveIssue,
			EnforceIssueTitle:          enforceIssueTitle,
			AutoEnableBackup:           autoEnableBackup,
			SkipBackupErrors:           skipBackupErrors,
			PostgresDatabaseTenantMode: postgresDatabaseTenantMode,
		}, paths); err != nil {
			diags = append(diags, diag.FromErr(err)...)
			return diags
		}
	}

	if d.HasChange("databases") {
		if diag := updateDatabasesInProject(ctx, d, c, d.Id()); diag != nil {
			diags = append(diags, diag...)
			return diags
		}
	}

	if d.HasChange("members") {
		if diag := updateMembersInProject(ctx, d, c, d.Id()); diag != nil {
			diags = append(diags, diag...)
			return diags
		}
	}

	diag := resourceProjectRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	projectName := d.Id()

	project, err := c.GetProject(ctx, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	resp := setProject(ctx, c, d, project)
	tflog.Debug(ctx, "[read project] read project finished", map[string]interface{}{
		"project": project.Name,
	})

	return resp
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	projectName := d.Id()

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if err := c.DeleteProject(ctx, projectName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func updateMembersInProject(ctx context.Context, d *schema.ResourceData, client api.Client, projectName string) diag.Diagnostics {
	memberSet, ok := d.Get("members").(*schema.Set)
	if !ok {
		return nil
	}

	iamPolicy := &v1pb.IamPolicy{}
	existProjectOwner := false

	for _, m := range memberSet.List() {
		rawMember := m.(map[string]interface{})
		expressions := []string{}

		if condition, ok := rawMember["condition"].(*schema.Set); ok {
			if condition.Len() > 1 {
				return diag.Errorf("should only set one condition")
			}
			if condition.Len() == 1 && condition.List()[0] != nil {
				rawCondition := condition.List()[0].(map[string]interface{})
				if database, ok := rawCondition["database"].(string); ok && database != "" {
					expressions = append(expressions, fmt.Sprintf(`resource.database == "%s"`, database))
				}
				if schema, ok := rawCondition["schema"].(string); ok {
					expressions = append(expressions, fmt.Sprintf(`resource.schema == "%s"`, schema))
				}
				if tables, ok := rawCondition["tables"].(*schema.Set); ok && tables.Len() > 0 {
					tableList := []string{}
					for _, table := range tables.List() {
						tableList = append(tableList, fmt.Sprintf(`"%s"`, table.(string)))
					}
					expressions = append(expressions, fmt.Sprintf(`resource.table in [%s]`, strings.Join(tableList, ",")))
				}
				if rowLimit, ok := rawCondition["row_limit"].(int); ok && rowLimit > 0 {
					expressions = append(expressions, fmt.Sprintf(`request.row_limit <= %d`, rowLimit))
				}
				if expire, ok := rawCondition["expire_timestamp"].(string); ok && expire != "" {
					formattedTime, err := time.Parse(time.RFC3339, expire)
					if err != nil {
						return diag.FromErr(errors.Wrapf(err, "invalid time: %v", expire))
					}
					expressions = append(expressions, fmt.Sprintf(`request.time < timestamp("%s")`, formattedTime.Format(time.RFC3339)))
				}
			}
		}

		member := rawMember["member"].(string)
		role := rawMember["role"].(string)
		if role == "roles/projectOwner" {
			existProjectOwner = true
		}

		if err := internal.ValidateMemberBinding(member); err != nil {
			return diag.FromErr(err)
		}
		if !strings.HasPrefix(role, internal.RoleNamePrefix) {
			return diag.Errorf("invalid role format")
		}

		iamPolicy.Bindings = append(iamPolicy.Bindings, &v1pb.Binding{
			Members: []string{member},
			Role:    role,
			Condition: &expr.Expr{
				Expression: strings.Join(expressions, " && "),
			},
		})
	}

	if len(iamPolicy.Bindings) > 0 {
		if !existProjectOwner {
			return diag.Errorf("require at least 1 member with roles/projectOwner role")
		}

		if _, err := client.SetProjectIAMPolicy(ctx, projectName, &v1pb.SetIamPolicyRequest{
			Policy: iamPolicy,
			Etag:   iamPolicy.Etag,
		}); err != nil {
			return diag.Errorf("failed to update iam for project %s with error: %v", projectName, err.Error())
		}
	}
	return nil
}

const batchSize = 100

func updateDatabasesInProject(ctx context.Context, d *schema.ResourceData, client api.Client, projectName string) diag.Diagnostics {
	databases, err := client.ListDatabase(ctx, projectName, "", true)
	if err != nil {
		return diag.Errorf("failed to list database with error: %v", err.Error())
	}
	existedDBMap := map[string]*v1pb.Database{}
	for _, db := range databases {
		existedDBMap[db.Name] = db
	}

	rawSet, ok := d.Get("databases").(*schema.Set)
	if !ok {
		return nil
	}
	updatedDBMap := map[string]*v1pb.Database{}
	batchTransferDatabases := []*v1pb.UpdateDatabaseRequest{}
	for _, raw := range rawSet.List() {
		dbName := raw.(string)
		if _, _, err := internal.GetInstanceDatabaseID(dbName); err != nil {
			return diag.Errorf("invalid database full name: %v", err.Error())
		}

		updatedDBMap[dbName] = &v1pb.Database{
			Name:    dbName,
			Project: projectName,
		}
		_, ok := existedDBMap[dbName]
		if !ok {
			// new assigned database
			batchTransferDatabases = append(batchTransferDatabases, &v1pb.UpdateDatabaseRequest{
				Database: updatedDBMap[dbName],
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"project"},
				},
			})
		} else {
			delete(existedDBMap, dbName)
		}
	}

	tflog.Debug(ctx, "[transfer databases] batch transfer databases to project", map[string]interface{}{
		"project":   projectName,
		"databases": len(batchTransferDatabases),
	})

	for i := 0; i < len(batchTransferDatabases); i += batchSize {
		end := i + batchSize
		if end > len(batchTransferDatabases) {
			end = len(batchTransferDatabases)
		}
		batch := batchTransferDatabases[i:end]
		startTime := time.Now()

		if _, err := client.BatchUpdateDatabases(ctx, &v1pb.BatchUpdateDatabasesRequest{
			Requests: batch,
			Parent:   "instances/-",
		}); err != nil {
			return diag.Errorf("failed to assign databases to project %s with error: %v", projectName, err.Error())
		}

		tflog.Debug(ctx, "[transfer databases]", map[string]interface{}{
			"count":   end + 1 - i,
			"project": projectName,
			"ms":      time.Since(startTime).Milliseconds(),
		})
	}

	if len(existedDBMap) > 0 {
		tflog.Debug(ctx, "[transfer databases] batch unassign databases", map[string]interface{}{
			"project":   projectName,
			"databases": len(existedDBMap),
		})

		startTime := time.Now()
		unassignDatabases := []*v1pb.UpdateDatabaseRequest{}
		for _, db := range existedDBMap {
			// move db to default project
			db.Project = defaultProj
			unassignDatabases = append(unassignDatabases, &v1pb.UpdateDatabaseRequest{
				Database: db,
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"project"},
				},
			})
		}
		if _, err := client.BatchUpdateDatabases(ctx, &v1pb.BatchUpdateDatabasesRequest{
			Requests: unassignDatabases,
			Parent:   "instances/-",
		}); err != nil {
			return diag.Errorf("failed to move databases to default project with error: %v", err.Error())
		}
		tflog.Debug(ctx, "[unassign databases]", map[string]interface{}{
			"count": len(unassignDatabases),
			"ms":    time.Since(startTime).Milliseconds(),
		})
	}

	return nil
}
