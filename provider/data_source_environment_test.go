package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccEnvironmentDataSource(t *testing.T) {
	identifier := "test"
	resourceName := fmt.Sprintf("bytebase_environment.%s", identifier)
	title := "test"
	order := 1

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			// get single environment test
			{
				Config: testAccCheckEnvironmentDataSource(
					testAccCheckEnvironmentResource(identifier, title, order),
					resourceName,
					identifier,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(fmt.Sprintf("data.%s", resourceName)),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s", resourceName), "title", title),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s", resourceName), "order", fmt.Sprintf("%d", order)),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s", resourceName), "environment_tier_policy", "PROTECTED"),
				),
			},
		},
	})
}

func TestAccEnvironmentDataSource_NotFound(t *testing.T) {
	identifier := "dev-notfound"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckEnvironmentDataSource(
					"",
					"",
					identifier,
				),
				ExpectError: regexp.MustCompile("Cannot found environment"),
			},
		},
	})
}

func testAccCheckEnvironmentDataSource(
	resource,
	dependsOn,
	identifier string,
) string {
	return fmt.Sprintf(`
	%s

	data "bytebase_environment" "%s" {
		resource_id = "%s"
		depends_on  = [
    		%s
  		]
	}
	`, resource, identifier, identifier, dependsOn)
}
