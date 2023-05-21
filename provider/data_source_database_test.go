package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccDatabaseDataSource(t *testing.T) {
	identifier := "default_db"
	instanceID := "dev-instance-with-db"
	environmentID := "dev-env"
	resourceName := fmt.Sprintf("data.bytebase_database.%s", identifier)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig(
					testAccCheckInstanceResource("dev_instance_with_db", instanceID, "Dev Instance", "POSTGRES", environmentID),
					identifier,
					"default",
					instanceID,
					"bytebase_instance.dev_instance_with_db",
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "labels.bb.environment", environmentID),
				),
			},
		},
	})
}

func testAccDatabaseConfig(
	resource,
	identifier,
	name,
	instance,
	dependsOn string,
) string {
	return fmt.Sprintf(`
	%s

	data "bytebase_database" "%s" {
		name = "%s"
		instance = "%s"
		depends_on = [
    		%s
  		]
	}
	`, resource, identifier, name, instance, dependsOn)
}
