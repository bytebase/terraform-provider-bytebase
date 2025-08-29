package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccRole(t *testing.T) {
	identifier := "test_role"
	resourceName := fmt.Sprintf("bytebase_role.%s", identifier)

	resourceID := "test-role"
	title := "Test Role"
	description := "A test role for terraform"

	titleUpdated := "Updated Test Role"
	descriptionUpdated := "An updated test role for terraform"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoleDestroy,
		Steps: []resource.TestStep{
			// resource create
			{
				Config: testAccCheckRoleResource(identifier, resourceID, title, description, []string{"bb.permission.database.query", "bb.permission.database.export"}),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_id", resourceID),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("roles/%s", resourceID)),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "2"),
				),
			},
			// resource update with more permissions
			{
				Config: testAccCheckRoleResource(identifier, resourceID, titleUpdated, descriptionUpdated, []string{"bb.permission.database.query", "bb.permission.database.export", "bb.permission.database.create"}),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_id", resourceID),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("roles/%s", resourceID)),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "3"),
				),
			},
			// resource update with fewer permissions
			{
				Config: testAccCheckRoleResource(identifier, resourceID, titleUpdated, descriptionUpdated, []string{"bb.permission.database.query"}),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_id", resourceID),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("roles/%s", resourceID)),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
				),
			},
		},
	})
}

func TestAccRole_InvalidInput(t *testing.T) {
	identifier := "invalid_role"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoleDestroy,
		Steps: []resource.TestStep{
			// Empty resource_id
			{
				Config: fmt.Sprintf(`
resource "bytebase_role" "%s" {
	resource_id = ""
	title       = "Test Role"
	description = "Description"
	permissions = ["bb.permission.database.query"]
}
`, identifier),
				ExpectError: regexp.MustCompile(`(expected "resource_id" to not be an empty string|invalid value for resource_id)`),
			},
			// Empty title
			{
				Config: fmt.Sprintf(`
resource "bytebase_role" "%s" {
	resource_id = "test-role"
	title       = ""
	description = "Description"
	permissions = ["bb.permission.database.query"]
}
`, identifier),
				ExpectError: regexp.MustCompile(`expected "title" to not be an empty string`),
			},
			// No permissions
			{
				Config: fmt.Sprintf(`
resource "bytebase_role" "%s" {
	resource_id = "test-role"
	title       = "Test Role"
	description = "Description"
	permissions = []
}
`, identifier),
				ExpectError: regexp.MustCompile(`(expected permissions to have at least \(1\)|Not enough list items|Attribute permissions requires 1 item minimum)`),
			},
			// Invalid permission format (not starting with bb.)
			{
				Config: fmt.Sprintf(`
resource "bytebase_role" "%s" {
	resource_id = "test-role"
	title       = "Test Role"
	description = "Description"
	permissions = ["invalid.permission"]
}
`, identifier),
				ExpectError: regexp.MustCompile(`(Permissions should start with "bb\." prefix|permission should start with "bb\." prefix)`),
			},
		},
	})
}

func testAccCheckRoleResource(identifier, resourceID, title, description string, permissions []string) string {
	permissionsStr := ""
	for _, p := range permissions {
		permissionsStr += fmt.Sprintf("\"%s\", ", p)
	}

	return fmt.Sprintf(`
resource "bytebase_role" "%s" {
	resource_id = "%s"
	title       = "%s"
	description = "%s"
	permissions = [%s]
}
`, identifier, resourceID, title, description, permissionsStr)
}

func testAccCheckRoleDestroy(s *terraform.State) error {
	c, ok := testAccProvider.Meta().(api.Client)
	if !ok {
		return errors.Errorf("cannot get the api client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_role" {
			continue
		}

		if err := c.DeleteRole(context.Background(), rs.Primary.ID); err != nil {
			return err
		}
	}

	return nil
}
