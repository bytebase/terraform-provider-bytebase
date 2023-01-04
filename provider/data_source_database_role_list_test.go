package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccDatabaseRoleListDataSource(t *testing.T) {
	instanceName := "test-instance"
	roleName := "test_role"
	outputName := "role_list"
	resourceName := fmt.Sprintf("data.bytebase_database_role_list.%s", outputName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				%s

				data "bytebase_database_role_list" "%s" {
					instance    = bytebase_instance.%s.resource_id
					environment = bytebase_instance.%s.environment
				}
				`, mockInstanceResource(instanceName), outputName, instanceName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "roles.#", "0"),
				),
			},
			{
				Config: fmt.Sprintf(`
				%s

				resource "bytebase_database_role" "%s" {
					name        = "%s"
					instance    = bytebase_instance.%s.resource_id
					environment = bytebase_instance.%s.environment

					attribute {}
				}

				data "bytebase_database_role_list" "%s" {
					instance = bytebase_instance.%s.resource_id
					environment = bytebase_instance.%s.environment

					depends_on = [
    					bytebase_database_role.%s
  					]
				}
				`, mockInstanceResource(instanceName), roleName, roleName, instanceName, instanceName, outputName, instanceName, instanceName, roleName),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "roles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "roles.0.name", roleName),
					resource.TestCheckResourceAttr(resourceName, "roles.0.connection_limit", "-1"),
					resource.TestCheckResourceAttr(resourceName, "roles.0.valid_until", ""),
				),
			},
		},
	})
}
