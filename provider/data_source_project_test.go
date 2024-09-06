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
	key := "BYT"

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
					testAccCheckProjectResource(identifier, resourceID, title, key),
					resourceName,
					identifier,
					resourceID,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(fmt.Sprintf("data.%s", resourceName)),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "key", key),
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
				ExpectError: regexp.MustCompile("Cannot found project"),
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
