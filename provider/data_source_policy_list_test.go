package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccPolicyListDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			internal.GetTestStepForDataSourceList(
				"",
				"",
				"bytebase_policy_list",
				"before",
				"policies",
				0,
			),
			internal.GetTestStepForDataSourceList(
				testAccCheckPolicyResource(
					"backup_plan",
					"environments/test",
					getBackupPlanPolicy(string(api.BackupPlanScheduleDaily), 999),
					api.PolicyTypeBackupPlan,
				),
				"bytebase_policy.backup_plan",
				"bytebase_policy_list",
				"after",
				"policies",
				1,
			),
		},
	})
}
