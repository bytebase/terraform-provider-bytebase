package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "The service account data source.",
		ReadContext: dataSourceServiceAccountRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s", internal.ServiceAccountNamePrefix),
				),
				Description: "The service account name in serviceAccounts/{email} format.",
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The display title of the service account.",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The service account email.",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The service account state.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The timestamp when the service account was created.",
			},
		},
	}
}

func dataSourceServiceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	name := d.Get("name").(string)

	sa, err := c.GetServiceAccount(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(sa.Name)

	return setServiceAccount(d, sa)
}

func setServiceAccount(d *schema.ResourceData, sa *v1pb.ServiceAccount) diag.Diagnostics {
	if err := d.Set("name", sa.Name); err != nil {
		return diag.Errorf("cannot set name for service account: %s", err.Error())
	}
	if err := d.Set("title", sa.Title); err != nil {
		return diag.Errorf("cannot set title for service account: %s", err.Error())
	}
	if err := d.Set("email", sa.Email); err != nil {
		return diag.Errorf("cannot set email for service account: %s", err.Error())
	}
	if err := d.Set("state", sa.State.String()); err != nil {
		return diag.Errorf("cannot set state for service account: %s", err.Error())
	}
	if sa.CreateTime != nil {
		if err := d.Set("create_time", sa.CreateTime.AsTime().UTC().Format(time.RFC3339)); err != nil {
			return diag.Errorf("cannot set create_time for service account: %s", err.Error())
		}
	}
	return nil
}
