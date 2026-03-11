package provider

import (
	"context"
	"fmt"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description:   "The service account resource.",
		ReadContext:   resourceServiceAccountRead,
		DeleteContext: resourceServiceAccountDelete,
		CreateContext: resourceServiceAccountCreate,
		UpdateContext: resourceServiceAccountUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"parent": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The parent resource. Format: projects/{project} for project-level, workspaces/{workspace id} for workspace-level. Defaults to the workspace if not specified.",
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s%s$", internal.WorkspaceNamePrefix, internal.ResourceIDPattern),
					fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern),
				),
			},
			"service_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The ID for the service account, which becomes part of the email.",
			},
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The display title of the service account.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The service account name in serviceAccounts/{email} format.",
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

func resourceServiceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	fullName := d.Id()
	sa, err := c.GetServiceAccount(ctx, fullName)
	if err != nil {
		if internal.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Resource %s not found, removing from state", fullName))
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return setServiceAccount(d, sa)
}

func resourceServiceAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	parent := internal.ResolveWorkspaceParent(d.Get("parent").(string), c.GetWorkspaceName())
	if err := d.Set("parent", parent); err != nil {
		return diag.Errorf("cannot set parent: %s", err.Error())
	}
	serviceAccountID := d.Get("service_account_id").(string)
	title := d.Get("title").(string)

	// Check if the service account already exists by listing and matching the ID.
	existedSA, err := findServiceAccountByID(ctx, c, serviceAccountID)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("find service account %s failed with error: %v", serviceAccountID, err))
	}

	var diags diag.Diagnostics
	if existedSA != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Service account already exists",
			Detail:   fmt.Sprintf("Service account %s already exists, try to exec the update operation", existedSA.Name),
		})

		if existedSA.State == v1pb.State_DELETED {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Service account is deleted",
				Detail:   fmt.Sprintf("Service account %s already deleted, try to undelete", existedSA.Name),
			})
			if _, err := c.UndeleteServiceAccount(ctx, existedSA.Name); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to undelete service account",
					Detail:   fmt.Sprintf("Undelete service account %s failed, error: %v", existedSA.Name, err),
				})
				return diags
			}
		}

		if title != existedSA.Title {
			if _, err := c.UpdateServiceAccount(ctx, &v1pb.ServiceAccount{
				Name:  existedSA.Name,
				Title: title,
			}, []string{"title"}); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to update service account",
					Detail:   fmt.Sprintf("Update service account %s failed, error: %v", existedSA.Name, err),
				})
				return diags
			}
		}
		d.SetId(existedSA.Name)
	} else {
		sa, err := c.CreateServiceAccount(ctx, parent, serviceAccountID, &v1pb.ServiceAccount{
			Title: title,
		})
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(sa.Name)
	}

	diag := resourceServiceAccountRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func findServiceAccountByID(ctx context.Context, c api.Client, serviceAccountID string) (*v1pb.ServiceAccount, error) {
	// The name is constructed as serviceAccounts/{id}@service.bytebase.com.
	name := fmt.Sprintf("%s%s@service.bytebase.com", internal.ServiceAccountNamePrefix, serviceAccountID)
	sa, err := c.GetServiceAccount(ctx, name)
	if err != nil {
		return nil, err
	}
	return sa, nil
}

func resourceServiceAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	name := d.Id()

	existedSA, err := c.GetServiceAccount(ctx, name)
	if err != nil {
		return diag.Errorf("get service account %s failed with error: %v", name, err)
	}

	var diags diag.Diagnostics
	if existedSA.State == v1pb.State_DELETED {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Service account is deleted",
			Detail:   fmt.Sprintf("Service account %s already deleted, try to undelete", name),
		})
		if _, err := c.UndeleteServiceAccount(ctx, name); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to undelete service account",
				Detail:   fmt.Sprintf("Undelete service account %s failed, error: %v", name, err),
			})
			return diags
		}
	}

	paths := []string{}
	if d.HasChange("title") {
		paths = append(paths, "title")
	}

	if len(paths) > 0 {
		title := d.Get("title").(string)
		if _, err := c.UpdateServiceAccount(ctx, &v1pb.ServiceAccount{
			Name:  existedSA.Name,
			Title: title,
		}, paths); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update service account",
				Detail:   fmt.Sprintf("Update service account %s failed, error: %v", name, err),
			})
			return diags
		}
	}

	diag := resourceServiceAccountRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceServiceAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	return internal.ResourceDelete(ctx, d, c.DeleteServiceAccount)
}
