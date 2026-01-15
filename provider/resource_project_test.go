package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccProject(t *testing.T) {
	identifier := "new_project"
	resourceName := fmt.Sprintf("bytebase_project.%s", identifier)

	resourceID := "test-project"
	title := "test project"
	titleUpdated := fmt.Sprintf("%s-updated", title)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			// resource create
			{
				Config: testAccCheckProjectResource(identifier, resourceID, title),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", title),
				),
			},
			// resource updated
			{
				Config: testAccCheckProjectResource(identifier, resourceID, titleUpdated),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
				),
			},
		},
	})
}

func TestAccProjectWithSettings(t *testing.T) {
	identifier := "project_with_settings"
	resourceName := fmt.Sprintf("bytebase_project.%s", identifier)

	resourceID := "test-project-settings"
	title := "test project with settings"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			// resource create with settings
			{
				Config: testAccCheckProjectResourceWithSettings(identifier, resourceID, title),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "enforce_sql_review", "true"),
					resource.TestCheckResourceAttr(resourceName, "require_issue_approval", "true"),
					resource.TestCheckResourceAttr(resourceName, "require_plan_check_no_error", "false"),
					resource.TestCheckResourceAttr(resourceName, "allow_request_role", "true"),
					resource.TestCheckResourceAttr(resourceName, "force_issue_labels", "true"),
					resource.TestCheckResourceAttr(resourceName, "issue_labels.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "labels.environment", "test"),
					resource.TestCheckResourceAttr(resourceName, "labels.team", "platform"),
				),
			},
			// resource update settings
			{
				Config: testAccCheckProjectResourceWithSettingsUpdated(identifier, resourceID, title),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enforce_sql_review", "false"),
					resource.TestCheckResourceAttr(resourceName, "require_issue_approval", "false"),
					resource.TestCheckResourceAttr(resourceName, "issue_labels.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "1"),
				),
			},
		},
	})
}

func testAccCheckProjectDestroy(s *terraform.State) error {
	c, ok := testAccProvider.Meta().(api.Client)
	if !ok {
		return errors.Errorf("cannot get the api client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_project" {
			continue
		}

		if err := c.DeleteProject(context.Background(), rs.Primary.ID); err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckProjectResource(identifier, resourceID, title string) string {
	return fmt.Sprintf(`
	resource "bytebase_project" "%s" {
		resource_id    = "%s"
		title          = "%s"
	}
	`, identifier, resourceID, title)
}

func testAccCheckProjectResourceWithSettings(identifier, resourceID, title string) string {
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
	`, identifier, resourceID, title)
}

func testAccCheckProjectResourceWithSettingsUpdated(identifier, resourceID, title string) string {
	return fmt.Sprintf(`
	resource "bytebase_project" "%s" {
		resource_id    = "%s"
		title          = "%s"

		enforce_sql_review          = false
		require_issue_approval      = false
		require_plan_check_no_error = true
		allow_request_role          = false
		force_issue_labels          = false

		issue_labels {
			value = "urgent"
			color = "#FFA500"
			group = "priority"
		}

		labels = {
			owner = "dba-team"
		}
	}
	`, identifier, resourceID, title)
}
