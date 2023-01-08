package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccInstanceRoleDataSource(t *testing.T) {
	roleName := "test_role"
	instanceName := "test-instance"
	resourceName := fmt.Sprintf("data.bytebase_instance_role.%s", roleName)

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

				resource "bytebase_instance_role" "%s" {
					name        = "%s"
					instance    = bytebase_instance.%s.resource_id
					environment = bytebase_instance.%s.environment

					attribute {}
				}

				data "bytebase_instance_role" "%s" {
					name        = bytebase_instance_role.%s.name
					instance    = bytebase_instance.%s.resource_id
					environment = bytebase_instance.%s.environment
				}
				`, mockInstanceResource(instanceName), roleName, roleName, instanceName, instanceName, roleName, roleName, instanceName, instanceName),
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

func TestAccInstanceRoleDataSource_RoleNotFound(t *testing.T) {
	roleName := "test_role"
	instanceName := "test-instance"

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

				data "bytebase_instance_role" "%s" {
					name        = "%s"
					instance    = bytebase_instance.%s.resource_id
					environment = bytebase_instance.%s.environment
				}
				`, mockInstanceResource(instanceName), roleName, roleName, instanceName, instanceName),
				ExpectError: regexp.MustCompile("Cannot found role with ID"),
			},
		},
	})
}

func TestAccInstanceRoleDataSource_InstanceNotFound(t *testing.T) {
	roleName := "test_role"
	instanceID := "test-instance"
	environmentID := "test-environment"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				data "bytebase_instance_role" "%s" {
					name        = "%s"
					instance    = "%s"
					environment = "%s"
				}
				`, roleName, roleName, instanceID, environmentID),
				ExpectError: regexp.MustCompile("Cannot found role with ID"),
			},
		},
	})
}
