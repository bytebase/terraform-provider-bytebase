package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccEnvironment(t *testing.T) {
	identifier := "test_env"
	resourceName := fmt.Sprintf("bytebase_environment.%s", identifier)

	resourceID := "test-environment"
	title := "Test Environment"
	order := 0
	titleUpdated := fmt.Sprintf("%s Updated", title)
	orderUpdated := 0

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			// resource create
			{
				Config: testAccCheckEnvironmentResource(identifier, resourceID, title, 1, 0.341176, 0.2, order, false),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_id", resourceID),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "color.0.red", "1"),
					resource.TestCheckResourceAttr(resourceName, "color.0.green", "0.341176"),
					resource.TestCheckResourceAttr(resourceName, "color.0.blue", "0.2"),
					resource.TestCheckResourceAttr(resourceName, "order", fmt.Sprintf("%d", order)),
					resource.TestCheckResourceAttr(resourceName, "protected", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("environments/%s", resourceID)),
				),
			},
			// resource update
			{
				Config: testAccCheckEnvironmentResource(identifier, resourceID, titleUpdated, 0.2, 1, 0.341176, orderUpdated, true),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_id", resourceID),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "color.0.red", "0.2"),
					resource.TestCheckResourceAttr(resourceName, "color.0.green", "1"),
					resource.TestCheckResourceAttr(resourceName, "color.0.blue", "0.341176"),
					resource.TestCheckResourceAttr(resourceName, "order", fmt.Sprintf("%d", orderUpdated)),
					resource.TestCheckResourceAttr(resourceName, "protected", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("environments/%s", resourceID)),
				),
			},
		},
	})
}

func TestAccEnvironment_InvalidInput(t *testing.T) {
	identifier := "invalid_env"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			// Invalid environment name (empty)
			{
				Config: fmt.Sprintf(`
resource "bytebase_environment" "%s" {
	resource_id = "test-env"
	title       = ""
	order       = 0
}
`, identifier),
				ExpectError: regexp.MustCompile(`expected "title" to not be an empty string`),
			},
			// Invalid resource_id (empty)
			{
				Config: fmt.Sprintf(`
resource "bytebase_environment" "%s" {
	resource_id = ""
	title       = "Test Environment"
	order       = 0
}
`, identifier),
				ExpectError: regexp.MustCompile(`(expected "resource_id" to not be an empty string|invalid value for resource_id)`),
			},
			// Invalid order (negative)
			{
				Config: fmt.Sprintf(`
resource "bytebase_environment" "%s" {
	resource_id = "test-env"
	title       = "Test Environment"
	order       = -1
}
`, identifier),
				ExpectError: regexp.MustCompile(`expected order to be at least \(0\)`),
			},
		},
	})
}

func testAccCheckEnvironmentResource(identifier, resourceID, title string, red, green, blue float64, order int, protected bool) string {
	return fmt.Sprintf(`
resource "bytebase_environment" "%s" {
	resource_id = "%s"
	title       = "%s"
	color {
		red   = %f
		green = %f
		blue  = %f
	}
	order       = %d
	protected   = %t
}
`, identifier, resourceID, title, red, green, blue, order, protected)
}

func testAccCheckEnvironmentDestroy(s *terraform.State) error {
	c, ok := testAccProvider.Meta().(api.Client)
	if !ok {
		return errors.Errorf("cannot get the api client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_environment" {
			continue
		}

		// Environment deletion is handled differently
		// We just check that it no longer exists
		_, _, envList, err := internal.FindEnvironment(context.Background(), c, rs.Primary.ID)
		if err != nil {
			if strings.Contains(err.Error(), "cannot found the environment") {
				continue
			}
			return err
		}

		for _, env := range envList {
			if env.Name == rs.Primary.ID {
				return errors.Errorf("environment %s still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}
