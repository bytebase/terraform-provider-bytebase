package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceReviewConfig() *schema.Resource {
	return &schema.Resource{
		Description: "The review config data source.",
		ReadContext: dataSourceReviewConfigRead,
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique resource id for the review config.",
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The title for the review config.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enable the SQL review config",
			},
			"resources": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Resources using the config. We support attach the review config for environments or projects with format {resurce}/{resource id}. For example, environments/test, projects/sample.",
			},
			"rules": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The SQL review rules.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The rule unique type. Check https://www.bytebase.com/docs/sql-review/review-rules for all rules",
						},
						"engine": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The rule for the database engine.",
						},
						"level": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The rule level.",
						},
						"payload": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The payload for the rule.",
						},
						"comment": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The comment for the rule.",
						},
					},
				},
				Set: reviewRuleHash,
			},
		},
	}
}

func dataSourceReviewConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	reviewName := fmt.Sprintf("%s%s", internal.ReviewConfigNamePrefix, d.Get("resource_id").(string))

	review, err := c.GetReviewConfig(ctx, reviewName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(review.Name)

	return setReviewConfig(d, review)
}
