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

func TestAccInstance(t *testing.T) {
	identifier := "new_instance"
	resourceName := fmt.Sprintf("bytebase_instance.%s", identifier)

	name := "dev instance"
	engine := "POSTGRES"
	host := "127.0.0.1"
	environment := "dev"
	nameUpdated := fmt.Sprintf("%s-updated", name)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			// resource create
			{
				Config: testAccCheckInstanceConfigBasic(identifier, name, engine, host, environment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "host", host),
					resource.TestCheckResourceAttr(resourceName, "environment", environment),
				),
			},
			// resource updated
			{
				Config: testAccCheckInstanceConfigBasic(identifier, nameUpdated, engine, host, environment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", nameUpdated),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "host", host),
					resource.TestCheckResourceAttr(resourceName, "environment", environment),
				),
			},
		},
	})
}

func TestAccInstance_InvalidInput(t *testing.T) {
	identifier := "another_instance"
	engine := "POSTGRES"
	name := "dev instance"
	host := "127.0.0.1"
	environment := "dev"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			// Invalid instance name
			{
				Config:      testAccCheckInstanceConfigBasic(identifier, "", engine, host, environment),
				ExpectError: regexp.MustCompile("not be an empty string"),
			},
			// Invalid engine
			{
				Config:      testAccCheckInstanceConfigBasic(identifier, name, "engine", host, environment),
				ExpectError: regexp.MustCompile("expected engine to be one of"),
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

func testAccCheckInstanceConfigBasic(identifier, name, engine, host, env string) string {
	return fmt.Sprintf(`
	resource "bytebase_instance" "%s" {
		name = "%s"
		engine = "%s"
		host = "%s"
		environment = "%s"
	}
	`, identifier, name, engine, host, env)
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
