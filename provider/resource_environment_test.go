package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func TestAccEnvironment(t *testing.T) {
	identifier := "new_environment"
	resourceName := fmt.Sprintf("bytebase_environment.%s", identifier)

	name := "dev"
	order := 1
	nameUpdated := fmt.Sprintf("%s-updated", name)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			// resource create
			{
				Config: testAccCheckEnvironmentConfigBasic(identifier, name, order),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "order", fmt.Sprintf("%d", order)),
				),
			},
			// resource update
			{
				Config: testAccCheckEnvironmentConfigBasic(identifier, nameUpdated, order+1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", nameUpdated),
					resource.TestCheckResourceAttr(resourceName, "order", fmt.Sprintf("%d", order+1)),
				),
			},
		},
	})
}

func TestAccEnvironment_InvalidInput(t *testing.T) {
	identifier := "another_environment"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			// Invalid environment name
			{
				Config:      testAccCheckEnvironmentConfigBasic(identifier, "", 0),
				ExpectError: regexp.MustCompile("not be an empty string"),
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

		envID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		if err := c.DeleteEnvironment(envID); err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckEnvironmentConfigBasic(identifier, envName string, order int) string {
	return fmt.Sprintf(`
	resource "bytebase_environment" "%s" {
		name = "%s"
		order = %d
	}
	`, identifier, envName, order)
}

func testAccCheckEnvironmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return errors.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.Errorf("No environment set")
		}

		return nil
	}
}
