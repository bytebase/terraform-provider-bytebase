package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func dataSourceGroupList() *schema.Resource {
	return &schema.Resource{
		Description: "The group data source list.",
		ReadContext: dataSourceGroupListRead,
		Schema: map[string]*schema.Schema{
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
						"creator": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The group creator in users/{email} format.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The group create time in YYYY-MM-DDThh:mm:ss.000Z format",
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

	response, err := c.ListGroup(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	groups := make([]map[string]interface{}, 0)
	for _, group := range response.Groups {
		raw := make(map[string]interface{})
		raw["name"] = group.Name
		raw["title"] = group.Title
		raw["description"] = group.Description
		raw["creator"] = group.Creator
		raw["create_time"] = group.CreateTime.AsTime().UTC().Format(time.RFC3339)
		raw["source"] = group.Source

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
