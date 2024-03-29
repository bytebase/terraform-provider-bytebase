package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccRole(t *testing.T) {
	roleName := "test_role"

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

				resource "bytebase_instance_role" "test_role_1" {
					name        = "%s"
					instance    = bytebase_instance.%s.resource_id

					attribute {}
				}
				`, mockInstanceResource("instance-1"), roleName, "instance-1"),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_instance_role.test_role_1"),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_1", "name", roleName),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_1", "instance", "instance-1"),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_1", "connection_limit", "-1"),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_1", "valid_until", ""),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_1", "attribute.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`
				%s

				resource "bytebase_instance_role" "test_role_2" {
					name        = "%s"
					instance    = bytebase_instance.%s.resource_id

					connection_limit = 99
					valid_until = "2022-12-31T15:04:05+08:00"
					attribute {}
				}
				`, mockInstanceResource("instance-2"), roleName, "instance-2"),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_instance_role.test_role_2"),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_2", "name", roleName),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_2", "instance", "instance-2"),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_2", "connection_limit", "99"),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_2", "valid_until", "2022-12-31T15:04:05+08:00"),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_2", "attribute.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`
				%s

				resource "bytebase_instance_role" "test_role_3" {
					name        = "%s"
					instance    = bytebase_instance.%s.resource_id

					attribute {
						super_user  = true
						no_inherit  = true
						create_role = false
					}
				}
				`, mockInstanceResource("instance-3"), roleName, "instance-3"),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_instance_role.test_role_3"),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_3", "name", roleName),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_3", "instance", "instance-3"),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_3", "attribute.#", "1"),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_3", "attribute.0.super_user", "true"),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_3", "attribute.0.no_inherit", "true"),
					resource.TestCheckResourceAttr("bytebase_instance_role.test_role_3", "attribute.0.create_role", "false"),
				),
			},
		},
	})
}

func testAccCheckRoleDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(api.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_instance_role" {
			continue
		}

		instanceID, roleName, err := internal.GetInstanceRoleID(rs.Primary.ID)
		if err != nil {
			return err
		}

		if err := c.DeleteRole(context.Background(), instanceID, roleName); err != nil {
			return err
		}
	}

	return nil
}

func mockInstanceResource(resourceID string) string {
	return fmt.Sprintf(`
	resource "bytebase_instance" "%s" {
		resource_id = "%s"
		title       = "%s"
		engine = "POSTGRES"
		environment = "test"

		data_sources {
			title     = "admin data source"
			type     = "ADMIN"
			username = "bytebase"
			host = "127.0.0.1"
			port = 3306
		}
	}
	`, resourceID, resourceID, resourceID)
}
