package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

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
					"masking_policy",
					"instances/test-sample-instance/databases/employee",
					getMaskingPolicy("salary", "amount", v1pb.MaskingLevel_FULL),
					v1pb.PolicyType_MASKING,
				),
				"bytebase_policy.masking_policy",
				"bytebase_policy_list",
				"after",
				"policies",
				1,
			),
		},
	})
}
