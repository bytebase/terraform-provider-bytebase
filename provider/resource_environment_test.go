package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func TestAccEnvironment(t *testing.T) {
	envName := "dev"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckEnvironmentConfigBasic(envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists("bytebase_environment.new"),
				),
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

func testAccCheckEnvironmentConfigBasic(envName string) string {
	return fmt.Sprintf(`
	resource "bytebase_environment" "new" {
		name = "%s"
	}
	`, envName)
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
