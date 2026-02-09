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

func TestAccUser(t *testing.T) {
	identifier := "test_user"
	resourceName := fmt.Sprintf("bytebase_user.%s", identifier)

	email := "test-user@example.com"
	title := "Test User"
	phone := "+1234567890"
	password := "SecureP@ssw0rd!"

	titleUpdated := "Updated Test User"
	phoneUpdated := "+0987654321"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			// resource create
			{
				Config: testAccCheckUserResource(identifier, email, title, phone, password),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "phone", phone),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "mfa_enabled", "false"),
				),
			},
			// resource update
			{
				Config: testAccCheckUserResource(identifier, email, titleUpdated, phoneUpdated, password),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "phone", phoneUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
				),
			},
		},
	})
}

func TestAccUser_InvalidInput(t *testing.T) {
	identifier := "invalid_user"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			// Empty email
			{
				Config:      testAccCheckUserResource(identifier, "", "Test User", "+123", "password"),
				ExpectError: regexp.MustCompile(`expected "email" to not be an empty string`),
			},
			// Empty title
			{
				Config:      testAccCheckUserResource(identifier, "test@example.com", "", "+123", "password"),
				ExpectError: regexp.MustCompile(`expected "title" to not be an empty string`),
			},
		},
	})
}

func testAccCheckUserResource(identifier, email, title, phone, password string) string {
	return fmt.Sprintf(`
resource "bytebase_user" "%s" {
	email    = "%s"
	title    = "%s"
	phone    = "%s"
	password = "%s"
}
`, identifier, email, title, phone, password)
}

func testAccCheckUserDestroy(s *terraform.State) error {
	c, ok := testAccProvider.Meta().(api.Client)
	if !ok {
		return errors.Errorf("cannot get the api client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_user" {
			continue
		}

		if err := c.DeleteUser(context.Background(), rs.Primary.ID); err != nil {
			return err
		}
	}

	return nil
}
