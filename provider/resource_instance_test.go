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

func TestAccInstance(t *testing.T) {
	identifier := "new_instance"
	resourceName := fmt.Sprintf("bytebase_instance.%s", identifier)

	resourceID := "test-instance"
	title := "test instance"
	engine := "POSTGRES"
	environment := "environments/test"
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
				Config: testAccCheckInstanceResourceWithLabels(identifier, resourceID, title, engine, environment, "test", "platform"),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "environment", environment),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "labels.environment", "test"),
					resource.TestCheckResourceAttr(resourceName, "labels.team", "platform"),
					resource.TestCheckResourceAttr(resourceName, "data_sources.#", "1"),
				),
			},
			// resource updated
			{
				Config: testAccCheckInstanceResourceWithLabels(identifier, resourceID, titleUpdated, engine, environment, "prod", "database"),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "environment", environment),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "labels.environment", "prod"),
					resource.TestCheckResourceAttr(resourceName, "labels.team", "database"),
					resource.TestCheckResourceAttr(resourceName, "data_sources.#", "1"),
				),
			},
		},
	})
}

func TestAccProjectInstance(t *testing.T) {
	const (
		parent       = "projects/sample-project"
		resourceID   = "project-instance"
		instanceName = "projects/sample-project/instances/project-instance"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "bytebase_instance" "project" {
  parent      = %q
  resource_id = %q
  title       = "Project instance"
  engine      = "POSTGRES"
  environment = "environments/test"

  data_sources {
    id       = "admin"
    type     = "ADMIN"
    username = "bytebase"
    host     = "127.0.0.1"
    port     = "5432"
  }
}

data "bytebase_instance" "project" {
  parent      = %q
  resource_id = %q
  depends_on  = [bytebase_instance.project]
}

data "bytebase_instance_list" "project" {
  parent     = %q
  depends_on = [bytebase_instance.project]
}

data "bytebase_instance_list" "workspace" {
  depends_on = [bytebase_instance.project]
}
`, parent, resourceID, parent, resourceID, parent),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bytebase_instance.project", "parent", parent),
					resource.TestCheckResourceAttr("bytebase_instance.project", "name", instanceName),
					resource.TestCheckResourceAttr("data.bytebase_instance.project", "parent", parent),
					resource.TestCheckResourceAttr("data.bytebase_instance.project", "name", instanceName),
					resource.TestCheckResourceAttr("data.bytebase_instance_list.project", "instances.#", "1"),
					resource.TestCheckResourceAttr("data.bytebase_instance_list.project", "instances.0.parent", parent),
					resource.TestCheckResourceAttr("data.bytebase_instance_list.project", "instances.0.name", instanceName),
					testCheckInstanceListExcludesName("data.bytebase_instance_list.workspace", instanceName),
				),
			},
		},
	})
}

func testCheckInstanceListExcludesName(resourceName, excludedName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instanceList, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return errors.Errorf("cannot find %s", resourceName)
		}

		for attribute, value := range instanceList.Primary.Attributes {
			if strings.HasPrefix(attribute, "instances.") && strings.HasSuffix(attribute, ".name") && value == excludedName {
				return errors.Errorf("%s contains excluded instance %s", resourceName, excludedName)
			}
		}
		return nil
	}
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
				Config:      testAccCheckInstanceResource(identifier, "test-instance", "", engine, "environments/test"),
				ExpectError: regexp.MustCompile(`expected "title" to not be an empty string`),
			},
			// Invalid engine
			{
				Config:      testAccCheckInstanceResource(identifier, "test-instance", "test instance", "engine", "environments/test"),
				ExpectError: regexp.MustCompile(`expected engine to be one of`),
			},
			// Invalid data source
			{
				Config: `
				resource "bytebase_instance" "test_instance" {
					resource_id = "test-instance"
					engine      = "POSTGRES"
					title       = "test instance"
					environment = "environments/test"
					data_sources {
						id = "read-only data source"
						type  = "READ_ONLY"
						host  = "127.0.0.1"
						port  = "3306"
					}
				}
				`,
				ExpectError: regexp.MustCompile(`data source "ADMIN" is required`),
			},
			// Invalid data source
			{
				Config: `
				resource "bytebase_instance" "test_instance" {
					resource_id = "test-instance"
					engine      = "POSTGRES"
					title       = "test instance"
					environment = "environments/test"
					data_sources {
						id = "unknown data source"
						type  = "UNKNOWN"
						host  = "127.0.0.1"
						port  = 5432
					}
				}
				`,
				ExpectError: regexp.MustCompile(`expected data_sources.0.type to be one of`),
			},
			// Invalid data source
			{
				Config: `
				resource "bytebase_instance" "test_instance" {
					resource_id = "test-instance"
					engine      = "POSTGRES"
					title       = "test instance"
					environment = "environments/test"
					data_sources {
						id = "admin data source"
						type  = "ADMIN"
						host  = "127.0.0.1"
						port  = 5432
					}
					data_sources {
						id = "admin data source 2"
						type  = "ADMIN"
						host  = "127.0.0.1"
						port  = 5432
					}
				}
				`,
				ExpectError: regexp.MustCompile(`duplicate data source type ADMIN`),
			},
		},
	})
}

func testAccCheckInstanceDestroy(s *terraform.State) error {
	c, ok := testAccProvider.Meta().(api.Client)
	if !ok {
		return errors.Errorf("cannot get the api client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_instance" {
			continue
		}

		if err := c.DeleteInstance(context.Background(), rs.Primary.ID); err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckInstanceResource(identifier, id, name, engine, env string) string {
	return testAccCheckInstanceResourceWithLabels(identifier, id, name, engine, env, "", "")
}

func testAccCheckInstanceResourceWithLabels(identifier, id, name, engine, env, labelEnvironment, labelTeam string) string {
	labels := ""
	if labelEnvironment != "" || labelTeam != "" {
		labels = fmt.Sprintf(`
		labels = {
			environment = "%s"
			team        = "%s"
		}
`, labelEnvironment, labelTeam)
	}
	return fmt.Sprintf(`
	resource "bytebase_instance" "%s" {
		resource_id = "%s"
		title       = "%s"
		engine      = "%s"
		environment = "%s"
		%s

		data_sources {
			id       = "admin data source"
			type     = "ADMIN"
			username = "bytebase"
			host     = "127.0.0.1"
			port     = "3306"
		}
	}
	`, identifier, id, name, engine, env, labels)
}
