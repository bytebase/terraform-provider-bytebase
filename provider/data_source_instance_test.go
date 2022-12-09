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

	name := "dev-instance"
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
			// get single instance test
			{
				Config: testAccCheckInstanceDataSource(
					testAccCheckInstanceResource(identifier, name, engine, host, environment),
					resourceName,
					identifier,
					name,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(fmt.Sprintf("data.%s", resourceName)),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "host", host),
					resource.TestCheckResourceAttr(resourceName, "environment", environment),
				),
			},
		},
	})
}

func TestAccInstanceDataSource_NotFound(t *testing.T) {
	identifier := "dev_instance"
	name := "dev"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckInstanceDataSource(
					"",
					"",
					identifier,
					name,
				),
				ExpectError: regexp.MustCompile("Unable to get the instance"),
			},
		},
	})
}

func testAccCheckInstanceDataSource(
	resource,
	dependsOn,
	identifier,
	name string) string {
	return fmt.Sprintf(`
	%s

	data "bytebase_instance" "%s" {
		name = "%s"
		depends_on = [
    		%s
  		]
	}
	`, resource, identifier, name, dependsOn)
}
