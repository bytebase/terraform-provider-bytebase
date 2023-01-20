package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccPolicyDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPolicyDataSource(
					testAccCheckPolicyResource(
						"backup_plan",
						"dev",
						getBackupPlanPolicy(string(api.BackupPlanScheduleDaily), 999),
						api.PolicyTypeBackupPlan,
					),
					"backup_plan",
					"dev",
					"bytebase_policy.backup_plan",
					api.PolicyTypeBackupPlan,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("data.bytebase_policy.backup_plan"),
					resource.TestCheckResourceAttr("data.bytebase_policy.backup_plan", "type", string(api.PolicyTypeBackupPlan)),
					resource.TestCheckResourceAttr("data.bytebase_policy.backup_plan", "backup_plan_policy.#", "1"),
					resource.TestCheckResourceAttr("data.bytebase_policy.backup_plan", "backup_plan_policy.0.schedule", string(api.BackupPlanScheduleDaily)),
					resource.TestCheckResourceAttr("data.bytebase_policy.backup_plan", "backup_plan_policy.0.retention_duration", "999"),
				),
			},
		},
	})
}

func TestAccPolicyDataSource_NotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPolicyDataSource(
					"",
					"policy",
					"dev",
					"",
					api.PolicyTypeDeploymentApproval,
				),
				ExpectError: regexp.MustCompile(fmt.Sprintf("Cannot found policy environments/dev/policies/%s", api.PolicyTypeDeploymentApproval)),
			},
		},
	})
}

func testAccCheckPolicyDataSource(
	resource,
	identifier,
	environment,
	dependsOn string,
	pType api.PolicyType) string {
	return fmt.Sprintf(`
	%s

	data "bytebase_policy" "%s" {
		environment = "%s"
		type        = "%s"
		depends_on  = [
    		%s
  		]
	}
	`, resource, identifier, environment, pType, dependsOn)
}
