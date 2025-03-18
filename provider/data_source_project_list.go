package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceProjectList() *schema.Resource {
	return &schema.Resource{
		Description:        "The project data source list.",
		ReadWithoutTimeout: dataSourceProjectListRead,
		Schema: map[string]*schema.Schema{
			"show_deleted": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Including removed project in the response.",
			},
			"projects": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The project unique resource id.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The project full name in projects/{resource id} format.",
						},
						"title": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The project title.",
						},
						"allow_modify_statement": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allow modifying statement after issue is created.",
						},
						"auto_resolve_issue": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Enable auto resolve issue.",
						},
						"enforce_issue_title": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Enforce issue title created by user instead of generated by Bytebase.",
						},
						"auto_enable_backup": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to automatically enable backup.",
						},
						"skip_backup_errors": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to skip backup errors and continue the data migration.",
						},
						"postgres_database_tenant_mode": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to enable the database tenant mode for PostgreSQL. If enabled, the issue will be created with the pre-appended \"set role <db_owner>\" statement.",
						},
						"members":   getProjectMembersSchema(true),
						"databases": getDatabasesSchema(true),
					},
				},
			},
		},
	}
}

func dataSourceProjectListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	allProjects, err := c.ListProject(ctx, d.Get("show_deleted").(bool))
	if err != nil {
		return diag.FromErr(err)
	}

	projects := make([]map[string]interface{}, 0)
	for _, project := range allProjects {
		projectID, err := internal.GetProjectID(project.Name)
		if err != nil {
			return diag.FromErr(err)
		}

		proj := make(map[string]interface{})
		proj["resource_id"] = projectID
		proj["name"] = project.Name
		proj["title"] = project.Title
		proj["allow_modify_statement"] = project.AllowModifyStatement
		proj["auto_resolve_issue"] = project.AutoResolveIssue
		proj["enforce_issue_title"] = project.EnforceIssueTitle
		proj["auto_enable_backup"] = project.AutoEnableBackup
		proj["skip_backup_errors"] = project.AllowModifyStatement
		proj["postgres_database_tenant_mode"] = project.PostgresDatabaseTenantMode

		databases, err := c.ListDatabase(ctx, project.Name, "", false)
		if err != nil {
			return diag.FromErr(err)
		}

		databaseList := flattenDatabaseList(databases)
		proj["databases"] = databaseList

		iamPolicy, err := c.GetProjectIAMPolicy(ctx, project.Name)
		if err != nil {
			return diag.Errorf("failed to get project iam with error: %v", err)
		}
		memberList, err := flattenMemberList(iamPolicy)
		if err != nil {
			return diag.FromErr(err)
		}
		proj["members"] = schema.NewSet(memberHash, memberList)

		projects = append(projects, proj)
	}

	if err := d.Set("projects", projects); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
