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

func TestAccInstance(t *testing.T) {
	identifier := "new_instance"
	resourceName := fmt.Sprintf("bytebase_instance.%s", identifier)

	resourceID := "dev-instance"
	title := "dev instance"
	engine := "POSTGRES"
	environment := "dev"
	titleUpdated := fmt.Sprintf("%s-updated", title)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			// resource create
			{
				Config: testAccCheckInstanceResource(identifier, resourceID, title, engine, environment),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "environment", environment),
					resource.TestCheckResourceAttr(resourceName, "data_sources.#", "1"),
				),
			},
			// resource updated
			{
				Config: testAccCheckInstanceResource(identifier, resourceID, titleUpdated, engine, environment),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "environment", environment),
					resource.TestCheckResourceAttr(resourceName, "data_sources.#", "1"),
				),
			},
		},
	})
}

func TestAccInstance_InvalidInput(t *testing.T) {
	identifier := "another_instance"
	engine := "POSTGRES"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			// Invalid instance name
			{
				Config:      testAccCheckInstanceResource(identifier, "dev-instance", "", engine, "dev"),
				ExpectError: regexp.MustCompile("expected \"title\" to not be an empty string"),
			},
			// Invalid engine
			{
				Config:      testAccCheckInstanceResource(identifier, "dev-instance", "dev instance", "engine", "dev"),
				ExpectError: regexp.MustCompile("expected engine to be one of"),
			},
			// Invalid data source
			{
				Config: `
				resource "bytebase_instance" "dev_instance" {
					resource_id = "dev-instance"
					engine      = "POSTGRES"
					title       = "dev instance"
					environment = "dev"
					data_sources {
						title = "read-only data source"
						type  = "READ_ONLY"
						host  = "127.0.0.1"
						port  = "3306"
					}
				}
				`,
				ExpectError: regexp.MustCompile("data source \"ADMIN\" is required"),
			},
			// Invalid data source
			{
				Config: `
				resource "bytebase_instance" "dev_instance" {
					resource_id = "dev-instance"
					engine      = "POSTGRES"
					title       = "dev instance"
					environment = "dev"
					data_sources {
						title = "unknown data source"
						type  = "UNKNOWN"
						host  = "127.0.0.1"
						port  = 5432
					}
				}
				`,
				ExpectError: regexp.MustCompile("expected data_sources.0.type to be one of"),
			},
			// Invalid data source
			{
				Config: `
				resource "bytebase_instance" "dev_instance" {
					resource_id = "dev-instance"
					engine      = "POSTGRES"
					title       = "dev instance"
					environment = "dev"
					data_sources {
						title = "admin data source"
						type  = "ADMIN"
						host  = "127.0.0.1"
						port  = 5432
					}
					data_sources {
						title = "admin data source 2"
						type  = "ADMIN"
						host  = "127.0.0.1"
						port  = 5432
					}
				}
				`,
				ExpectError: regexp.MustCompile("duplicate data source type ADMIN"),
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

		instanceID, err := internal.GetInstanceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		if err := c.DeleteInstance(context.Background(), instanceID); err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckInstanceResource(identifier, id, name, engine, env string) string {
	return fmt.Sprintf(`
	resource "bytebase_instance" "%s" {
		resource_id = "%s"
		title       = "%s"
		engine      = "%s"
		environment = "%s"

		data_sources {
			title    = "admin data source"
			type     = "ADMIN"
			username = "bytebase"
			host     = "127.0.0.1"
			port     = "3306"
		}
	}
	`, identifier, id, name, engine, env)
}
