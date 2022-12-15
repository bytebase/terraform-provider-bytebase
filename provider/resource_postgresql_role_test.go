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

func TestAccPGRole(t *testing.T) {
	roleName := "test_role"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "bytebase_postgresql_role" "test_role_1" {
					name = "%s"
					instance_id = 1
					attribute {}
				}
				`, roleName),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_postgresql_role.test_role_1"),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_1", "name", roleName),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_1", "instance_id", "1"),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_1", "connection_limit", "-1"),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_1", "valid_until", ""),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_1", "attribute.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "bytebase_postgresql_role" "test_role_2" {
					name = "%s"
					instance_id = 2
					connection_limit = 99
					valid_until = "2022-12-31"
					attribute {}
				}
				`, roleName),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_postgresql_role.test_role_2"),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_2", "name", roleName),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_2", "instance_id", "2"),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_2", "connection_limit", "99"),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_2", "valid_until", "2022-12-31"),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_2", "attribute.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "bytebase_postgresql_role" "test_role_3" {
					name = "%s"
					instance_id = 3
					attribute {
						super_user  = true
						no_inherit  = true
						create_role = false
					}
				}
				`, roleName),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_postgresql_role.test_role_3"),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_3", "name", roleName),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_3", "instance_id", "3"),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_3", "attribute.#", "1"),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_3", "attribute.0.super_user", "true"),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_3", "attribute.0.no_inherit", "true"),
					resource.TestCheckResourceAttr("bytebase_postgresql_role.test_role_3", "attribute.0.create_role", "false"),
				),
			},
		},
	})
}

func testAccCheckPGRoleDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(api.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_postgresql_role" {
			continue
		}

		instanceID, name, err := parseRoleIdentifier(rs.Primary.ID)
		if err != nil {
			return err
		}

		if err := c.DeletePGRole(context.Background(), instanceID, name); err != nil {
			return err
		}
	}

	return nil
}
