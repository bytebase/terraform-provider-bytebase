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

				resource "bytebase_database_role" "test_role_1" {
					name = "%s"
					instance = bytebase_instance.%s.name
					attribute {}
				}
				`, mockInstanceResource("instance_1"), roleName, "instance_1"),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_database_role.test_role_1"),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_1", "name", roleName),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_1", "instance", "instance_1"),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_1", "connection_limit", "-1"),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_1", "valid_until", ""),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_1", "attribute.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`
				%s

				resource "bytebase_database_role" "test_role_2" {
					name = "%s"
					instance = bytebase_instance.%s.name
					connection_limit = 99
					valid_until = "2022-12-31"
					attribute {}
				}
				`, mockInstanceResource("instance_2"), roleName, "instance_2"),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_database_role.test_role_2"),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_2", "name", roleName),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_2", "instance", "instance_2"),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_2", "connection_limit", "99"),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_2", "valid_until", "2022-12-31"),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_2", "attribute.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`
				%s

				resource "bytebase_database_role" "test_role_3" {
					name = "%s"
					instance = bytebase_instance.%s.name
					attribute {
						super_user  = true
						no_inherit  = true
						create_role = false
					}
				}
				`, mockInstanceResource("instance_3"), roleName, "instance_3"),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_database_role.test_role_3"),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_3", "name", roleName),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_3", "instance", "instance_3"),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_3", "attribute.#", "1"),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_3", "attribute.0.super_user", "true"),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_3", "attribute.0.no_inherit", "true"),
					resource.TestCheckResourceAttr("bytebase_database_role.test_role_3", "attribute.0.create_role", "false"),
				),
			},
		},
	})
}

func testAccCheckRoleDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(api.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_database_role" {
			continue
		}

		instanceID, name, err := parseRoleIdentifier(rs.Primary.ID)
		if err != nil {
			return err
		}

		if err := c.DeleteRole(context.Background(), instanceID, name); err != nil {
			return err
		}
	}

	return nil
}

func mockInstanceResource(name string) string {
	return fmt.Sprintf(`
	resource "bytebase_instance" "%s" {
		name = "%s"
		engine = "POSTGRES"
		host = "127.0.0.1"
		port = 3306
		environment = "dev"

		data_source_list {
			name     = "admin data source"
			type     = "ADMIN"
			username = "bytebase"
		}
	}
	`, name, name)
}