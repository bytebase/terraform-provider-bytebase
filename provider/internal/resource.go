package internal

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// IsNotFoundError checks if the error is a 404 Not Found error.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error message contains status code 404
	return strings.Contains(err.Error(), "not_found")
}

// ResourceDeleteFunc is the func to delete the resource by name.
type ResourceDeleteFunc func(ctx context.Context, name string) error

// ResourceDelete wrap the delete func.
func ResourceDelete(ctx context.Context, d *schema.ResourceData, delete ResourceDeleteFunc) diag.Diagnostics {
	fullName := d.Id()

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if err := delete(ctx, fullName); err != nil {
		// Check if the resource was deleted outside of Terraform
		if !IsNotFoundError(err) {
			return diag.FromErr(err)
		}
	}

	d.SetId("")

	return diags
}
