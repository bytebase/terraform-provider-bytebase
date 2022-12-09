package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccEnvironment(t *testing.T) {
	identifier := "new_environment"
	resourceName := fmt.Sprintf("bytebase_environment.%s", identifier)

	name := "dev"
	order := 1
	nameUpdated := fmt.Sprintf("%s-updated", name)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			// resource create test
			{
				Config: testAccCheckEnvironmentResource(identifier, name, order),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "order", fmt.Sprintf("%d", order)),
				),
			},
			// resource update test
			{
				Config: testAccCheckEnvironmentResource(identifier, nameUpdated, order+1),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", nameUpdated),
					resource.TestCheckResourceAttr(resourceName, "order", fmt.Sprintf("%d", order+1)),
				),
			},
		},
	})
}

func TestAccEnvironment_InvalidInput(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			// Invalid environment order
			{
				Config: fmt.Sprintf(`
					resource "bytebase_environment" "new_env" {
						name = "new_env"
						environment_tier_policy  = "PROTECTED"
						pipeline_approval_policy = "MANUAL_APPROVAL_BY_WORKSPACE_OWNER_OR_DBA"
						backup_plan_policy       = "DAILY"
					}
				`),
				ExpectError: regexp.MustCompile("The argument \"order\" is required, but no definition was found"),
			},
			// Invalid environment name
			{
				Config: fmt.Sprintf(`
					resource "bytebase_environment" "new_env" {
						name = ""
						order = 1
						environment_tier_policy  = "PROTECTED"
						pipeline_approval_policy = "MANUAL_APPROVAL_BY_WORKSPACE_OWNER_OR_DBA"
						backup_plan_policy       = "DAILY"
					}
				`),
				ExpectError: regexp.MustCompile("expected \"name\" to not be an empty string"),
			},
			// Invalid policy
			{
				Config: fmt.Sprintf(`
					resource "bytebase_environment" "new_env" {
						name = "new_env"
						order = 1
						pipeline_approval_policy = "MANUAL_APPROVAL_BY_WORKSPACE_OWNER_OR_DBA"
						backup_plan_policy       = "DAILY"
					}
				`),
				ExpectError: regexp.MustCompile("The argument \"environment_tier_policy\" is required"),
			},
			// Invalid policy
			{
				Config: fmt.Sprintf(`
					resource "bytebase_environment" "new_env" {
						name = "new_env"
						order = 1
						environment_tier_policy = "UNKNOWN"
						pipeline_approval_policy = "MANUAL_APPROVAL_BY_WORKSPACE_OWNER_OR_DBA"
						backup_plan_policy       = "DAILY"
					}
				`),
				ExpectError: regexp.MustCompile("expected environment_tier_policy to be one of"),
			},
		},
	})
}

func testAccCheckEnvironmentDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(api.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_environment" {
			continue
		}

		envID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		if err := c.DeleteEnvironment(context.Background(), envID); err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckEnvironmentResource(identifier, envName string, order int) string {
	return fmt.Sprintf(`
	resource "bytebase_environment" "%s" {
		name = "%s"
		order = %d
		environment_tier_policy  = "PROTECTED"
		pipeline_approval_policy = "MANUAL_APPROVAL_BY_WORKSPACE_OWNER_OR_DBA"
		backup_plan_policy       = "DAILY"
	}
	`, identifier, envName, order)
}
