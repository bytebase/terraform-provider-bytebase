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

func dataSourceReviewConfigList() *schema.Resource {
	return &schema.Resource{
		Description: "The review config data source list.",
		ReadContext: dataSourceReviewConfigListRead,
		Schema: map[string]*schema.Schema{
			"review_configs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
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
							Type:        schema.TypeList,
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
						},
					},
				},
			},
		},
	}
}

func dataSourceReviewConfigListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	response, err := c.ListReviewConfig(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	reviews := make([]map[string]interface{}, 0)
	for _, review := range response.ReviewConfigs {
		raw := make(map[string]interface{})
		reviewID, err := internal.GetReviewConfigID(review.Name)
		if err != nil {
			return diag.Errorf("failed to parse id from review name %s with error: %v", review.Name, err.Error())
		}
		raw["resource_id"] = reviewID
		raw["title"] = review.Title
		raw["enabled"] = review.Enabled
		raw["resources"] = review.Resources
		raw["rules"] = flattenReviewRules(review.Rules)

		reviews = append(reviews, raw)
	}

	if err := d.Set("review_configs", reviews); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
