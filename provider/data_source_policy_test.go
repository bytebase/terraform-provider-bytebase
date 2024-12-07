package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

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
						"masking_policy",
						"instances/test-sample-instance/databases/employee",
						getMaskingPolicy("salary", "amount", v1pb.MaskingLevel_FULL),
						v1pb.PolicyType_MASKING,
					),
					"masking_policy",
					"instances/test-sample-instance/databases/employee",
					"bytebase_policy.masking_policy",
					v1pb.PolicyType_MASKING,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("data.bytebase_policy.masking_policy"),
					resource.TestCheckResourceAttr("data.bytebase_policy.masking_policy", "type", v1pb.PolicyType_MASKING.String()),
					resource.TestCheckResourceAttr("data.bytebase_policy.masking_policy", "masking_policy.#", "1"),
					resource.TestCheckResourceAttr("data.bytebase_policy.masking_policy", "masking_policy.0.mask_data.#", "1"),
					resource.TestCheckResourceAttr("data.bytebase_policy.masking_policy", "masking_policy.0.mask_data.0.table", "salary"),
					resource.TestCheckResourceAttr("data.bytebase_policy.masking_policy", "masking_policy.0.mask_data.0.column", "amount"),
					resource.TestCheckResourceAttr("data.bytebase_policy.masking_policy", "masking_policy.0.mask_data.0.masking_level", v1pb.MaskingLevel_FULL.String()),
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
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPolicyDataSource(
					"",
					"policy",
					"instances/test-sample-instance/databases/employee",
					"",
					v1pb.PolicyType_MASKING,
				),
				ExpectError: regexp.MustCompile("Cannot found policy instances/test-sample-instance/databases/employee/policies/MASKING"),
			},
		},
	})
}

func testAccCheckPolicyDataSource(
	resource,
	identifier,
	parent,
	dependsOn string,
	pType v1pb.PolicyType) string {
	return fmt.Sprintf(`
	%s

	data "bytebase_policy" "%s" {
		parent     = "%s"
		type       = "%s"
		depends_on = [
    		%s
  		]
	}
	`, resource, identifier, parent, pType.String(), dependsOn)
}
