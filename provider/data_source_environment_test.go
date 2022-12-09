package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccEnvironmentDataSource(t *testing.T) {
	identifier := "dev"
	resourceName := fmt.Sprintf("bytebase_environment.%s", identifier)
	name := "dev"
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
					testAccCheckEnvironmentResource(identifier, name, order),
					resourceName,
					identifier,
					name,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(fmt.Sprintf("data.%s", resourceName)),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s", resourceName), "name", name),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s", resourceName), "order", fmt.Sprintf("%d", order)),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s", resourceName), "environment_tier_policy", "PROTECTED"),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s", resourceName), "pipeline_approval_policy", "MANUAL_APPROVAL_BY_WORKSPACE_OWNER_OR_DBA"),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s", resourceName), "backup_plan_policy", "DAILY"),
				),
			},
		},
	})
}

func TestAccEnvironmentDataSource_NotFound(t *testing.T) {
	identifier := "dev"
	name := "dev"

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
					name,
				),
				ExpectError: regexp.MustCompile("Unable to get the environment"),
			},
		},
	})
}

func testAccCheckEnvironmentDataSource(
	resource,
	dependsOn,
	identifier,
	envName string,
) string {
	return fmt.Sprintf(`
	%s

	data "bytebase_environment" "%s" {
		name = "%s"
		depends_on = [
    		%s
  		]
	}
	`, resource, identifier, envName, dependsOn)
}
