package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceGroupList() *schema.Resource {
	return &schema.Resource{
		Description: "The group data source list.",
		ReadContext: dataSourceGroupListRead,
		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern),
				),
				Description: "The project fullname in projects/{id} format.",
			},
			"query": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter groups by title or email with wildcard",
			},
			"groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The group name in groups/{email} format.",
						},
						"title": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The group title.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The group description.",
						},
						"source": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Source means where the group comes from. For now we support Entra ID SCIM sync, so the source could be Entra ID.",
						},
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The group email.",
						},
						"members": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "The members in the group.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"member": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The member in users/{email} format.",
									},
									"role": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The member's role in the group.",
									},
								},
							},
							Set: memberHash,
						},
					},
				},
			},
		},
	}
}

func dataSourceGroupListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	response, err := c.ListGroup(ctx, &api.GroupFilter{
		Query:   d.Get("query").(string),
		Project: d.Get("project").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	groups := make([]map[string]interface{}, 0)
	for _, group := range response {
		raw := make(map[string]interface{})
		raw["name"] = group.Name
		raw["title"] = group.Title
		raw["description"] = group.Description
		raw["source"] = group.Source
		raw["email"] = group.Email

		memberList := []interface{}{}
		for _, member := range group.Members {
			rawMember := map[string]interface{}{}
			rawMember["member"] = member.Member
			rawMember["role"] = member.Role.String()
			memberList = append(memberList, rawMember)
		}
		raw["members"] = schema.NewSet(memberHash, memberList)

		groups = append(groups, raw)
	}

	if err := d.Set("groups", groups); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
