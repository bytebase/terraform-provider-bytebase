package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"

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
						"masking_exemption_policy",
						"projects/project-sample",
						getMaskingExemptionPolicy("instances/test-sample-instance/databases/employee", "salary", "amount"),
						v1pb.PolicyType_MASKING_EXEMPTION,
					),
					"masking_exemption_policy",
					"projects/project-sample",
					"bytebase_policy.masking_exemption_policy",
					v1pb.PolicyType_MASKING_EXEMPTION,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("data.bytebase_policy.masking_exemption_policy"),
					resource.TestCheckResourceAttr("data.bytebase_policy.masking_exemption_policy", "type", v1pb.PolicyType_MASKING_EXEMPTION.String()),
					resource.TestCheckResourceAttr("data.bytebase_policy.masking_exemption_policy", "masking_exemption_policy.#", "1"),
					resource.TestCheckResourceAttr("data.bytebase_policy.masking_exemption_policy", "masking_exemption_policy.0.exemptions.#", "1"),
					resource.TestCheckResourceAttr("data.bytebase_policy.masking_exemption_policy", "masking_exemption_policy.0.exemptions.0.table", "salary"),
					resource.TestCheckResourceAttr("data.bytebase_policy.masking_exemption_policy", "masking_exemption_policy.0.exemptions.0.columns.0", "amount"),
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
					"projects/project-sample",
					"",
					v1pb.PolicyType_MASKING_EXEMPTION,
				),
				ExpectError: regexp.MustCompile(`Cannot found policy projects/project-sample/policies/MASKING_EXEMPTION`),
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
