package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccProject(t *testing.T) {
	identifier := "new_project"
	resourceName := fmt.Sprintf("bytebase_project.%s", identifier)

	resourceID := "test-project"
	title := "test project"
	key := "BYT"
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
				Config: testAccCheckProjectResource(identifier, resourceID, title, key, api.ProjectWorkflowUI, api.ProjectSchemaVersionSemantic, api.ProjectSchemaChangeDDL),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "workflow", string(api.ProjectWorkflowUI)),
					resource.TestCheckResourceAttr(resourceName, "schema_version", string(api.ProjectSchemaVersionSemantic)),
					resource.TestCheckResourceAttr(resourceName, "schema_change", string(api.ProjectSchemaChangeDDL)),
				),
			},
			// resource updated
			{
				Config: testAccCheckProjectResource(identifier, resourceID, titleUpdated, key, api.ProjectWorkflowVCS, api.ProjectSchemaVersionSemantic, api.ProjectSchemaChangeSDL),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "workflow", string(api.ProjectWorkflowVCS)),
					resource.TestCheckResourceAttr(resourceName, "schema_version", string(api.ProjectSchemaVersionSemantic)),
					resource.TestCheckResourceAttr(resourceName, "schema_change", string(api.ProjectSchemaChangeSDL)),
				),
			},
		},
	})
}

func testAccCheckProjectDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(api.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_project" {
			continue
		}

		projectID, err := internal.GetProjectID(rs.Primary.ID)
		if err != nil {
			return err
		}

		if err := c.DeleteProject(context.Background(), projectID); err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckProjectResource(identifier, resourceID, title, key string, workflow api.ProjectWorkflow, schemaVersion api.ProjectSchemaVersion, schemaChange api.ProjectSchemaChange) string {
	return fmt.Sprintf(`
	resource "bytebase_project" "%s" {
		resource_id    = "%s"
		title          = "%s"
		key            = "%s"
		workflow       = "%s"
		schema_version = "%s"
		schema_change  = "%s"
		tenant_mode    = "TENANT_MODE_DISABLED"
		visibility     = "VISIBILITY_PUBLIC"
	}
	`, identifier, resourceID, title, key, workflow, schemaVersion, schemaChange)
}
