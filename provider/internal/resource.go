package internal

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

// isNotFoundError checks if the error is a 404 Not Found error
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error message contains status code 404
	return strings.Contains(err.Error(), "status: 404") ||
		strings.Contains(err.Error(), "status: "+string(rune(http.StatusNotFound)))
}

// ResourceRead read the resource, and will clear the state if the resource not exist.
// Once the state is cleared, the terraform can exec the creation.
func ResourceRead(read schema.ReadContextFunc) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		c := m.(api.Client)
		var diags diag.Diagnostics

		fullName := d.Id()
		if err := c.CheckResourceExist(ctx, fullName); err != nil {
			// Check if the resource was deleted outside of Terraform
			if isNotFoundError(err) {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Resource not found",
					Detail:   fmt.Sprintf("Resource %s not found, removing from state then create a new one", fullName),
				})
				tflog.Warn(ctx, fmt.Sprintf("Resource %s not found, removing from state", fullName))
				// Remove from state to trigger recreation on next apply
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}

		return read(ctx, d, m)
	}
}

// ResourceDelete force delete the resource.
func ResourceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	fullName := d.Id()

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if err := c.DeleteResource(ctx, fullName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
