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

func TestAccGroup(t *testing.T) {
	identifier := "test_group"
	resourceName := fmt.Sprintf("bytebase_group.%s", identifier)

	email := "test-group@example.com"
	title := "Test Group"
	description := "A test group for terraform"

	titleUpdated := "Updated Test Group"
	descriptionUpdated := "An updated test group for terraform"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			// resource create with single member - simplified version
			{
				Config: fmt.Sprintf(`
resource "bytebase_user" "test_owner" {
	email    = "owner@example.com"
	title    = "Test Owner"
	password = "test_password_123"
	type     = "USER"
}

resource "bytebase_group" "%s" {
	email       = "%s"
	title       = "%s"
	description = "%s"
	
	members {
		member = bytebase_user.test_owner.name
		role   = "OWNER"
	}
}
`, identifier, email, title, description),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("groups/%s", email)),
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
				),
			},
			// resource update with multiple members
			{
				Config: fmt.Sprintf(`
resource "bytebase_user" "test_owner" {
	email    = "owner@example.com"
	title    = "Test Owner"
	password = "test_password_123"
	type     = "USER"
}

resource "bytebase_user" "test_member1" {
	email    = "member1@example.com"
	title    = "Test Member 1"
	password = "test_password_123"
	type     = "USER"
}

resource "bytebase_user" "test_member2" {
	email    = "member2@example.com"
	title    = "Test Member 2"
	password = "test_password_123"
	type     = "USER"
}

resource "bytebase_group" "%s" {
	email       = "%s"
	title       = "%s"
	description = "%s"
	
	members {
		member = bytebase_user.test_owner.name
		role   = "OWNER"
	}
	members {
		member = bytebase_user.test_member1.name
		role   = "MEMBER"
	}
	members {
		member = bytebase_user.test_member2.name
		role   = "MEMBER"
	}
}
`, identifier, email, titleUpdated, descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("groups/%s", email)),
					resource.TestCheckResourceAttr(resourceName, "members.#", "3"),
				),
			},
			// resource update - change member roles
			{
				Config: fmt.Sprintf(`
resource "bytebase_user" "test_owner" {
	email    = "owner@example.com"
	title    = "Test Owner"
	password = "test_password_123"
	type     = "USER"
}

resource "bytebase_user" "test_member1" {
	email    = "member1@example.com"
	title    = "Test Member 1"
	password = "test_password_123"
	type     = "USER"
}

resource "bytebase_user" "test_member2" {
	email    = "member2@example.com"
	title    = "Test Member 2"
	password = "test_password_123"
	type     = "USER"
}

resource "bytebase_group" "%s" {
	email       = "%s"
	title       = "%s"
	description = "%s"
	
	members {
		member = bytebase_user.test_owner.name
		role   = "OWNER"
	}
	members {
		member = bytebase_user.test_member1.name
		role   = "OWNER"
	}
	members {
		member = bytebase_user.test_member2.name
		role   = "MEMBER"
	}
}
`, identifier, email, titleUpdated, descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "members.#", "3"),
				),
			},
			// resource update - remove members
			{
				Config: fmt.Sprintf(`
resource "bytebase_user" "test_owner" {
	email    = "owner@example.com"
	title    = "Test Owner"
	password = "test_password_123"
	type     = "USER"
}

resource "bytebase_user" "test_member1" {
	email    = "member1@example.com"
	title    = "Test Member 1"
	password = "test_password_123"
	type     = "USER"
}

resource "bytebase_group" "%s" {
	email       = "%s"
	title       = "%s"
	description = "%s"
	
	members {
		member = bytebase_user.test_owner.name
		role   = "OWNER"
	}
	members {
		member = bytebase_user.test_member1.name
		role   = "OWNER"
	}
}
`, identifier, email, titleUpdated, descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "members.#", "2"),
				),
			},
		},
	})
}

func TestAccGroup_InvalidInput(t *testing.T) {
	identifier := "invalid_group"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			// Empty email
			{
				Config: fmt.Sprintf(`
resource "bytebase_user" "test_user_%s" {
	email    = "test@example.com"
	title    = "Test User"
	password = "test_password"
	type     = "USER"
}

resource "bytebase_group" "%s" {
	email       = ""
	title       = "Test Group"
	description = "Description"
	members {
		member = bytebase_user.test_user_%s.name
		role   = "OWNER"
	}
}
`, identifier, identifier, identifier),
				ExpectError: regexp.MustCompile(`expected "email" to not be an empty string`),
			},
			// Empty title
			{
				Config: fmt.Sprintf(`
resource "bytebase_user" "test_user2_%s" {
	email    = "test2@example.com"
	title    = "Test User 2"
	password = "test_password"
	type     = "USER"
}

resource "bytebase_group" "%s" {
	email       = "group@example.com"
	title       = ""
	description = "Description"
	members {
		member = bytebase_user.test_user2_%s.name
		role   = "OWNER"
	}
}
`, identifier, identifier, identifier),
				ExpectError: regexp.MustCompile(`expected "title" to not be an empty string`),
			},
			// No members
			{
				Config: fmt.Sprintf(`
resource "bytebase_group" "%s" {
	email       = "group@example.com"
	title       = "Test Group"
	description = "Description"
}
`, identifier),
				ExpectError: regexp.MustCompile(`(expected members to have at least|At least 1 "members" blocks are required|Missing required argument)`),
			},
			// Invalid member format
			{
				Config: fmt.Sprintf(`
resource "bytebase_group" "%s" {
	email       = "group@example.com"
	title       = "Test Group"
	description = "Description"
	members {
		member = "invalid-member-format"
		role   = "OWNER"
	}
}
`, identifier),
				ExpectError: regexp.MustCompile(`(expected value of member to match regular expression|Resource id not match|doesn't must any patterns)`),
			},
			// Invalid role
			{
				Config: fmt.Sprintf(`
resource "bytebase_group" "%s" {
	email       = "group@example.com"
	title       = "Test Group"
	description = "Description"
	members {
		member = "users/test@example.com"
		role   = "INVALID_ROLE"
	}
}
`, identifier),
				ExpectError: regexp.MustCompile(`(expected role to be one of|expected members\.0\.role to be one of)`),
			},
		},
	})
}

func testAccCheckGroupDestroy(s *terraform.State) error {
	c, ok := testAccProvider.Meta().(api.Client)
	if !ok {
		return errors.Errorf("cannot get the api client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_group" {
			continue
		}

		if err := c.DeleteGroup(context.Background(), rs.Primary.ID); err != nil {
			return err
		}
	}

	return nil
}
