package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccDatabaseRoleDataSource(t *testing.T) {
	roleName := "test_role"
	instanceName := "test_instance"
	resourceName := fmt.Sprintf("data.bytebase_database_role.%s", roleName)

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

				resource "bytebase_database_role" "%s" {
					name = "%s"
					instance = bytebase_instance.%s.name
					attribute {}
				}

				data "bytebase_database_role" "%s" {
					name     = bytebase_database_role.%s.name
					instance = bytebase_instance.%s.name
				}
				`, mockInstanceResource(instanceName), roleName, roleName, instanceName, roleName, roleName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "instance", instanceName),
					resource.TestCheckResourceAttr(resourceName, "connection_limit", "-1"),
					resource.TestCheckResourceAttr(resourceName, "valid_until", ""),
				),
			},
		},
	})
}

func TestAccDatabaseRoleDataSource_RoleNotFound(t *testing.T) {
	roleName := "test_role"
	instanceName := "test_instance"

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

				data "bytebase_database_role" "%s" {
					name     = "%s"
					instance = bytebase_instance.%s.name
				}
				`, mockInstanceResource(instanceName), roleName, roleName, instanceName),
				ExpectError: regexp.MustCompile("Cannot found role with ID"),
			},
		},
	})
}

func TestAccDatabaseRoleDataSource_InstanceNotFound(t *testing.T) {
	roleName := "test_role"
	instanceName := "test_instance"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				data "bytebase_database_role" "%s" {
					name     = "%s"
					instance = "%s"
				}
				`, roleName, roleName, instanceName),
				ExpectError: regexp.MustCompile("Cannot find the instance"),
			},
		},
	})
}
