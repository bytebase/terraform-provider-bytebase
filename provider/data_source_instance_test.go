package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccInstanceDataSource(t *testing.T) {
	identifier := "new_instance"
	resourceName := fmt.Sprintf("bytebase_instance.%s", identifier)

	resourceID := "test-instance"
	title := "test instance"
	engine := "POSTGRES"
	environment := "environments/test"
	dataSourceName := fmt.Sprintf("data.%s", resourceName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			// get single instance test
			{
				Config: testAccCheckInstanceDataSource(
					testAccCheckInstanceResourceWithLabels(identifier, resourceID, title, engine, environment, "test", "platform"),
					resourceName,
					identifier,
					resourceID,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(dataSourceName),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "environment", environment),
					resource.TestCheckResourceAttr(dataSourceName, "labels.%", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "labels.environment", "test"),
					resource.TestCheckResourceAttr(dataSourceName, "labels.team", "platform"),
				),
			},
		},
	})
}

func TestAccInstanceDataSource_NotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckInstanceDataSource(
					"",
					"",
					"test_instance",
					"mock-id",
				),
				ExpectError: regexp.MustCompile(`Cannot found instance`),
			},
		},
	})
}

func testAccCheckInstanceDataSource(
	resource,
	dependsOn,
	identifier,
	resourceID string) string {
	return fmt.Sprintf(`
	%s

	data "bytebase_instance" "%s" {
		resource_id = "%s"
		depends_on  = [
    		%s
  		]
	}
	`, resource, identifier, resourceID, dependsOn)
}
