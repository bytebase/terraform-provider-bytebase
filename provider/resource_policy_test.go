package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccPolicy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPolicyResource(
					"backup_plan",
					"dev",
					getBackupPlanPolicy(string(api.BackupPlanScheduleDaily), 999),
					api.PolicyTypeBackupPlan,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_policy.backup_plan"),
					resource.TestCheckResourceAttr("bytebase_policy.backup_plan", "type", string(api.PolicyTypeBackupPlan)),
					resource.TestCheckResourceAttr("bytebase_policy.backup_plan", "backup_plan_policy.#", "1"),
					resource.TestCheckResourceAttr("bytebase_policy.backup_plan", "backup_plan_policy.0.schedule", string(api.BackupPlanScheduleDaily)),
					resource.TestCheckResourceAttr("bytebase_policy.backup_plan", "backup_plan_policy.0.retention_duration", "999"),
				),
			},
			{
				Config: testAccCheckPolicyResource(
					"backup_plan",
					"dev",
					getBackupPlanPolicy(string(api.BackupPlanScheduleWeekly), 99),
					api.PolicyTypeBackupPlan,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_policy.backup_plan"),
					resource.TestCheckResourceAttr("bytebase_policy.backup_plan", "type", string(api.PolicyTypeBackupPlan)),
					resource.TestCheckResourceAttr("bytebase_policy.backup_plan", "backup_plan_policy.#", "1"),
					resource.TestCheckResourceAttr("bytebase_policy.backup_plan", "backup_plan_policy.0.schedule", string(api.BackupPlanScheduleWeekly)),
					resource.TestCheckResourceAttr("bytebase_policy.backup_plan", "backup_plan_policy.0.retention_duration", "99"),
				),
			},
			{
				Config: testAccCheckPolicyResource(
					"deployment_approval",
					"dev",
					getDeploymentApprovalPolicy(string(api.ApprovalStrategyAutomatic), []*api.DeploymentApprovalStrategy{
						{
							ApprovalGroup:    api.ApprovalGroupDBA,
							ApprovalStrategy: api.ApprovalStrategyAutomatic,
							DeploymentType:   api.DeploymentTypeDatabaseCreate,
						},
						{
							ApprovalGroup:    api.ApprovalGroupOwner,
							ApprovalStrategy: api.ApprovalStrategyAutomatic,
							DeploymentType:   api.DeploymentTypeDatabaseDDL,
						},
					}),
					api.PolicyTypeDeploymentApproval,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_policy.deployment_approval"),
					resource.TestCheckResourceAttr("bytebase_policy.deployment_approval", "type", string(api.PolicyTypeDeploymentApproval)),
					resource.TestCheckResourceAttr("bytebase_policy.deployment_approval", "deployment_approval_policy.#", "1"),
					resource.TestCheckResourceAttr("bytebase_policy.deployment_approval", "deployment_approval_policy.0.default_strategy", string(api.ApprovalStrategyAutomatic)),
					resource.TestCheckResourceAttr("bytebase_policy.deployment_approval", "deployment_approval_policy.0.deployment_approval_strategies.#", "2"),
					resource.TestCheckResourceAttr("bytebase_policy.deployment_approval", "deployment_approval_policy.0.deployment_approval_strategies.0.deployment_type", string(api.DeploymentTypeDatabaseCreate)),
					resource.TestCheckResourceAttr("bytebase_policy.deployment_approval", "deployment_approval_policy.0.deployment_approval_strategies.1.deployment_type", string(api.DeploymentTypeDatabaseDDL)),
				),
			},
		},
	})
}

func TestAccPolicy_InvalidInput(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPolicyResource(
					"backup_plan",
					"dev",
					getBackupPlanPolicy("daily", 999),
					api.PolicyTypeBackupPlan,
				),
				ExpectError: regexp.MustCompile("expected backup_plan_policy.0.schedule to be one of"),
			},
			{
				Config: testAccCheckPolicyResource(
					"deployment_approval",
					"dev",
					getDeploymentApprovalPolicy("unknown", []*api.DeploymentApprovalStrategy{
						{
							ApprovalGroup:    api.ApprovalGroupDBA,
							ApprovalStrategy: api.ApprovalStrategyAutomatic,
							DeploymentType:   api.DeploymentTypeDatabaseCreate,
						},
						{
							ApprovalGroup:    api.ApprovalGroupOwner,
							ApprovalStrategy: api.ApprovalStrategyAutomatic,
							DeploymentType:   api.DeploymentTypeDatabaseDDL,
						},
					}),
					api.PolicyTypeDeploymentApproval,
				),
				ExpectError: regexp.MustCompile("expected deployment_approval_policy.0.default_strategy to be one of"),
			},
		},
	})
}

func testAccCheckPolicyResource(identifier, environment, payload string, pType api.PolicyType) string {
	return fmt.Sprintf(`
	resource "bytebase_policy" "%s" {
		environment = "%s"
		type        = "%s"

		%s
	}
	`, identifier, environment, pType, payload)
}

func getBackupPlanPolicy(schedule string, duration int) string {
	return fmt.Sprintf(`
	backup_plan_policy {
		schedule           = "%s"
		retention_duration = %d
	}
	`, schedule, duration)
}

func getDeploymentApprovalPolicy(defaultStrategy string, strategies []*api.DeploymentApprovalStrategy) string {
	approvalStrategies := []string{}
	for _, strategy := range strategies {
		approvalStrategies = append(approvalStrategies, fmt.Sprintf(`
		deployment_approval_strategies {
			approval_group    = "%s"
			approval_strategy = "%s"
			deployment_type   = "%s"
		}
		`, strategy.ApprovalGroup, strategy.ApprovalStrategy, strategy.DeploymentType))
	}

	return fmt.Sprintf(`
	deployment_approval_policy {
		default_strategy           = "%s"
		%s
	}
	`, defaultStrategy, strings.Join(approvalStrategies, "\n"))
}

func testAccCheckPolicyDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(api.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_policy" {
			continue
		}

		find, err := internal.GetPolicyFindMessageByName(rs.Primary.ID)
		if err != nil {
			return err
		}

		if err := c.DeletePolicy(context.Background(), find); err != nil {
			return err
		}
	}

	return nil
}
