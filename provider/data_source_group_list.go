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
						"roles": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "The group's roles in the workspace level",
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

	workspaceIAM, err := c.GetWorkspaceIAMPolicy(ctx)
	if err != nil {
		return diag.Errorf("cannot get workspace IAM with error: %s", err.Error())
	}

	groups := make([]map[string]interface{}, 0)
	for _, group := range response.Groups {
		raw := make(map[string]interface{})
		raw["name"] = group.Name
		raw["title"] = group.Title
		raw["description"] = group.Description
		raw["source"] = group.Source

		memberList := []interface{}{}
		for _, member := range group.Members {
			rawMember := map[string]interface{}{}
			rawMember["member"] = member.Member
			rawMember["role"] = member.Role.String()
			memberList = append(memberList, rawMember)
		}
		raw["members"] = schema.NewSet(memberHash, memberList)

		groupEmail, err := internal.GetGroupEmail(group.Name)
		if err != nil {
			return diag.Errorf("failed to parse group email: %v", err)
		}
		raw["roles"] = getRolesInIAM(workspaceIAM, fmt.Sprintf("group:%s", groupEmail))
		groups = append(groups, raw)
	}

	if err := d.Set("groups", groups); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
