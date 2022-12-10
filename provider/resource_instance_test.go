package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccInstance(t *testing.T) {
	identifier := "new_instance"
	resourceName := fmt.Sprintf("bytebase_instance.%s", identifier)

	name := "dev-instance"
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
				Config: testAccCheckInstanceResource(identifier, name, engine, host, environment),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "host", host),
					resource.TestCheckResourceAttr(resourceName, "environment", environment),
					resource.TestCheckResourceAttr(resourceName, "data_source_list.#", "1"),
				),
			},
			// resource updated
			{
				Config: testAccCheckInstanceResource(identifier, nameUpdated, engine, host, environment),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", nameUpdated),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "host", host),
					resource.TestCheckResourceAttr(resourceName, "environment", environment),
					resource.TestCheckResourceAttr(resourceName, "data_source_list.#", "1"),
				),
			},
		},
	})
}

func TestAccInstance_InvalidInput(t *testing.T) {
	identifier := "another_instance"
	engine := "POSTGRES"
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
				Config:      testAccCheckInstanceResource(identifier, "", engine, host, environment),
				ExpectError: regexp.MustCompile("expected \"name\" to not be an empty string"),
			},
			// Invalid engine
			{
				Config:      testAccCheckInstanceResource(identifier, "dev-instance", "engine", host, environment),
				ExpectError: regexp.MustCompile("expected engine to be one of"),
			},
			// Invalid data source
			{
				Config: `
				resource "bytebase_instance" "dev_instance" {
					name        = "dev"
					engine      = "POSTGRES"
					host        = "127.0.0.1"
					port        = 5432
					environment = "dev"
					data_source_list {
						name     = "read-only data source"
						type     = "RO"
					}
				}
				`,
				ExpectError: regexp.MustCompile("data source \"ADMIN\" is required"),
			},
			// Invalid data source
			{
				Config: `
				resource "bytebase_instance" "dev_instance" {
					name        = "dev"
					engine      = "POSTGRES"
					host        = "127.0.0.1"
					port        = 5432
					environment = "dev"
					data_source_list {
						name     = "unknown data source"
						type     = "UNKNOWN"
					}
				}
				`,
				ExpectError: regexp.MustCompile("expected data_source_list.0.type to be one of"),
			},
			// Invalid data source
			{
				Config: `
				resource "bytebase_instance" "dev_instance" {
					name        = "dev"
					engine      = "POSTGRES"
					host        = "127.0.0.1"
					port        = 5432
					environment = "dev"
					data_source_list {
						name     = "admin data source"
						type     = "ADMIN"
					}
					data_source_list {
						name     = "admin data source 2"
						type     = "ADMIN"
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

		instanceID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		if err := c.DeleteInstance(context.Background(), instanceID); err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckInstanceResource(identifier, name, engine, host, env string) string {
	return fmt.Sprintf(`
	resource "bytebase_instance" "%s" {
		name = "%s"
		engine = "%s"
		host = "%s"
		port = 3306
		environment = "%s"

		data_source_list {
			name     = "admin data source"
			type     = "ADMIN"
			username = "bytebase"
		}
	}
	`, identifier, name, engine, host, env)
}
