package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccProjectDataSource(t *testing.T) {
	identifier := "new_project"
	resourceName := fmt.Sprintf("bytebase_project.%s", identifier)

	resourceID := "test-project"
	title := "test project"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			// get single project test
			{
				Config: testAccCheckProjectDataSource(
					testAccCheckProjectResource(identifier, resourceID, title),
					resourceName,
					identifier,
					resourceID,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(fmt.Sprintf("data.%s", resourceName)),
					resource.TestCheckResourceAttr(resourceName, "title", title),
				),
			},
		},
	})
}

func TestAccProjectDataSourceWithSettings(t *testing.T) {
	identifier := "project_with_settings"
	dataSourceName := fmt.Sprintf("data.bytebase_project.%s", identifier)

	resourceID := "test-project-ds"
	title := "test project data source"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckProjectDataSourceWithSettings(
					identifier,
					resourceID,
					title,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "title", title),
					resource.TestCheckResourceAttr(dataSourceName, "enforce_sql_review", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "require_issue_approval", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "allow_request_role", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "force_issue_labels", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "issue_labels.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "labels.%", "2"),
				),
			},
		},
	})
}

func TestAccProjectDataSource_NotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckProjectDataSource(
					"",
					"",
					"mock_instance",
					"mock-id",
				),
				ExpectError: regexp.MustCompile(`Cannot found project`),
			},
		},
	})
}

func testAccCheckProjectDataSource(
	resource,
	dependsOn,
	identifier,
	resourceID string) string {
	return fmt.Sprintf(`
	%s

	data "bytebase_project" "%s" {
		resource_id = "%s"
		depends_on = [
    		%s
  		]
	}
	`, resource, identifier, resourceID, dependsOn)
}

func testAccCheckProjectDataSourceWithSettings(identifier, resourceID, title string) string {
	return fmt.Sprintf(`
	resource "bytebase_project" "%s" {
		resource_id    = "%s"
		title          = "%s"

		enforce_sql_review          = true
		require_issue_approval      = true
		require_plan_check_no_error = false
		allow_request_role          = true
		force_issue_labels          = true

		issue_labels {
			value = "bug"
			color = "#FF0000"
			group = "type"
		}
		issue_labels {
			value = "feature"
			color = "#00FF00"
			group = "type"
		}

		labels = {
			environment = "test"
			team        = "platform"
		}
	}

	data "bytebase_project" "%s" {
		resource_id = "%s"
		depends_on = [
			bytebase_project.%s
		]
	}
	`, identifier, resourceID, title, identifier, resourceID, identifier)
}
