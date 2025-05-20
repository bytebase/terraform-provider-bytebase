package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceDatabaseGroupList() *schema.Resource {
	return &schema.Resource{
		Description: "The database group data source list.",
		ReadContext: dataSourceDatabaseGroupListRead,
		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					regexp.MustCompile(fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern)),
				),
				Description: "The project fullname in projects/{id} format.",
			},
			"database_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The database group fullname in projects/{id}/databaseGroups/{id} format.",
						},
						"title": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The database group title.",
						},
						"condition": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The database group condition.",
						},
					},
				},
			},
		},
	}
}

func dataSourceDatabaseGroupListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	projectName := d.Get("project").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	response, err := c.ListDatabaseGroup(ctx, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	groups := []map[string]interface{}{}
	for _, group := range response.DatabaseGroups {
		raw := make(map[string]interface{})
		raw["name"] = group.Name
		raw["title"] = group.Title
		raw["condition"] = group.DatabaseExpr.Expression
		groups = append(groups, raw)
	}

	if err := d.Set("database_groups", groups); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
