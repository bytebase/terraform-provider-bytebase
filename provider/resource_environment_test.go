package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccEnvironment(t *testing.T) {
	identifier := "new-environment"
	resourceName := fmt.Sprintf("bytebase_environment.%s", identifier)

	title := "test"
	order := 1
	titleUpdated := fmt.Sprintf("%supdated", title)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			// resource create test
			{
				Config: testAccCheckEnvironmentResource(identifier, title, order),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "order", fmt.Sprintf("%d", order)),
				),
			},
			// resource update test
			{
				Config: testAccCheckEnvironmentResource(identifier, titleUpdated, order+1),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "order", fmt.Sprintf("%d", order+1)),
				),
			},
		},
	})
}

func TestAccEnvironment_InvalidInput(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			// Invalid environment order
			{
				Config: `
					resource "bytebase_environment" "new_env" {
						resource_id              = "test"
						title                    = "Test"
						environment_tier_policy  = "PROTECTED"
					}
				`,
				ExpectError: regexp.MustCompile("The argument \"order\" is required, but no definition was found"),
			},
			// Invalid environment name
			{
				Config: `
					resource "bytebase_environment" "new_env" {
						resource_id              = "test"
						title                    = ""
						order                    = 1
						environment_tier_policy  = "PROTECTED"
					}
				`,
				ExpectError: regexp.MustCompile("environment title must matches"),
			},
			// Invalid policy
			{
				Config: `
					resource "bytebase_environment" "new_env" {
						resource_id             = "test"
						title                   = "Test"
						order                   = 1
						environment_tier_policy = "UNKNOWN"
					}
				`,
				ExpectError: regexp.MustCompile("expected environment_tier_policy to be one of"),
			},
		},
	})
}

func testAccCheckEnvironmentDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(api.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_environment" {
			continue
		}

		if err := c.DeleteEnvironment(context.Background(), rs.Primary.ID); err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckEnvironmentResource(identifier, envName string, order int) string {
	return fmt.Sprintf(`
	resource "bytebase_environment" "%s" {
		resource_id             = "%s"
		title                   = "%s"
		order                   = %d
		environment_tier_policy = "PROTECTED"
	}
	`, identifier, identifier, envName, order)
}
