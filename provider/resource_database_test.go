package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccDatabase(t *testing.T) {
	identifier := "default_db"
	instanceID := "dev-instance-with-db"
	environmentID := "dev-env"
	projectID := "dev-project"
	resourceName := fmt.Sprintf("bytebase_database.%s", identifier)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseResourceConfig(
					testAccCheckInstanceResource("dev_instance_with_db", instanceID, "Dev Instance", "POSTGRES", environmentID),
					identifier,
					"default",
					instanceID,
					projectID,
					fmt.Sprintf("%s-updated", environmentID),
					"bytebase_instance.dev_instance_with_db",
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "project", projectID),
				),
			},
		},
	})
}

func testAccDatabaseResourceConfig(
	resource,
	identifier,
	name,
	instance,
	project,
	environment,
	dependsOn string,
) string {
	return fmt.Sprintf(`
	%s

	resource "bytebase_database" "%s" {
		name = "%s"
		instance = "%s"
		project = "%s"
		labels = {
    		"bb.environment" = "%s"
  		}
		depends_on = [
    		%s
  		]
	}
	`, resource, identifier, name, instance, project, environment, dependsOn)
}
