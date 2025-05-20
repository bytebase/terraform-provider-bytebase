package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceDatabaseGroup() *schema.Resource {
	return &schema.Resource{
		Description: "The database group data source.",
		ReadContext: dataSourceDatabaseGroupRead,
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
					regexp.MustCompile(fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern)),
				),
				Description: "The project fullname in projects/{id} format.",
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

func dataSourceDatabaseGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	groupID := d.Get("resource_id").(string)
	projectName := d.Get("project").(string)
	groupName := fmt.Sprintf("%s/%s%s", projectName, internal.DatabaseGroupNamePrefix, groupID)

	group, err := c.GetDatabaseGroup(ctx, groupName, v1pb.DatabaseGroupView_DATABASE_GROUP_VIEW_FULL)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(group.Name)

	return setDatabaseGroup(d, group)
}
