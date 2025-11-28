package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceGroup() *schema.Resource {
	return &schema.Resource{
		Description: "The group data source.",
		ReadContext: dataSourceGroupRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s", internal.GroupNamePrefix),
				),
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
	}
}

func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	groupName := d.Get("name").(string)
	group, err := c.GetGroup(ctx, groupName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(group.Name)
	return setGroup(d, group)
}

func setGroup(d *schema.ResourceData, group *v1pb.Group) diag.Diagnostics {
	if err := d.Set("name", group.Name); err != nil {
		return diag.Errorf("cannot set name for group: %s", err.Error())
	}
	if err := d.Set("title", group.Title); err != nil {
		return diag.Errorf("cannot set title for group: %s", err.Error())
	}
	if err := d.Set("description", group.Description); err != nil {
		return diag.Errorf("cannot set description for group: %s", err.Error())
	}
	if err := d.Set("source", group.Source); err != nil {
		return diag.Errorf("cannot set source for group: %s", err.Error())
	}
	if err := d.Set("email", group.Email); err != nil {
		return diag.Errorf("cannot set email for group: %s", err.Error())
	}

	memberList := []interface{}{}
	for _, member := range group.Members {
		rawMember := map[string]interface{}{}
		rawMember["member"] = member.Member
		rawMember["role"] = member.Role.String()
		memberList = append(memberList, rawMember)
	}
	if err := d.Set("members", schema.NewSet(memberHash, memberList)); err != nil {
		return diag.Errorf("cannot set members for group: %s", err.Error())
	}

	return nil
}

func memberHash(rawMember interface{}) int {
	member := convertToV1Member(rawMember)
	return internal.ToHash(member)
}
