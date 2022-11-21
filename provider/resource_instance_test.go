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

func TestAccInstance(t *testing.T) {
	name := "dev instance"
	engine := "POSTGRES"
	host := "127.0.0.1"
	env := "dev"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckInstanceConfigBasic(name, engine, host, env),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("bytebase_instance.new"),
				),
			},
		},
	})
}

func testAccCheckInstanceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(api.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_instance" {
			continue
		}

		instanceID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		if err := c.DeleteInstance(instanceID); err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckInstanceConfigBasic(name, engine, host, env string) string {
	return fmt.Sprintf(`
	resource "bytebase_instance" "new" {
		name = "%s"
		engine = "%s"
		host = "%s"
		environment = "%s"
	}
	`, name, engine, host, env)
}

func testAccCheckInstanceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return errors.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.Errorf("No instance set")
		}

		return nil
	}
}
