package provider

import (
	"context"
	"fmt"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

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
					"masking_exemption_policy",
					"projects/project-sample",
					getMaskingExemptionPolicy("instances/test-sample-instance/databases/employee", "salary", "amount"),
					v1pb.PolicyType_MASKING_EXEMPTION,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_policy.masking_exemption_policy"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exemption_policy", "type", v1pb.PolicyType_MASKING_EXEMPTION.String()),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exemption_policy", "masking_exemption_policy.#", "1"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exemption_policy", "masking_exemption_policy.0.exemptions.#", "1"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exemption_policy", "masking_exemption_policy.0.exemptions.0.table", "salary"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exemption_policy", "masking_exemption_policy.0.exemptions.0.columns.0", "amount"),
				),
			},
		},
	})
}

func testAccCheckPolicyResource(identifier, parent, payload string, pType v1pb.PolicyType) string {
	return fmt.Sprintf(`
	resource "bytebase_policy" "%s" {
		parent = "%s"
		type   = "%s"

		%s
	}
	`, identifier, parent, pType.String(), payload)
}

func getMaskingExemptionPolicy(database, table, column string) string {
	return fmt.Sprintf(`
	masking_exemption_policy {
		exemptions {
			database      = "%s"
			table         = "%s"
			columns       = ["%s"]
			members       = ["user:ed@bytebase.com"]
		}
	}
	`, database, table, column)
}

func getQueryDataPolicy(disableExport bool, maxResultSize, maxResultRows, timeoutInSeconds int) string {
	return fmt.Sprintf(`
	query_data_policy {
		disable_export       = %t
		maximum_result_size  = %d
		maximum_result_rows  = %d
		timeout_in_seconds   = %d
	}
	`, disableExport, maxResultSize, maxResultRows, timeoutInSeconds)
}

func TestAccPolicy_QueryData(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPolicyResource(
					"query_data_policy",
					"workspaces/-",
					getQueryDataPolicy(true, 1000000, 500, 60),
					v1pb.PolicyType_DATA_QUERY,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_policy.query_data_policy"),
					resource.TestCheckResourceAttr("bytebase_policy.query_data_policy", "type", v1pb.PolicyType_DATA_QUERY.String()),
					resource.TestCheckResourceAttr("bytebase_policy.query_data_policy", "query_data_policy.#", "1"),
					resource.TestCheckResourceAttr("bytebase_policy.query_data_policy", "query_data_policy.0.disable_export", "true"),
					resource.TestCheckResourceAttr("bytebase_policy.query_data_policy", "query_data_policy.0.maximum_result_size", "1000000"),
					resource.TestCheckResourceAttr("bytebase_policy.query_data_policy", "query_data_policy.0.maximum_result_rows", "500"),
					resource.TestCheckResourceAttr("bytebase_policy.query_data_policy", "query_data_policy.0.timeout_in_seconds", "60"),
				),
			},
			{
				Config: testAccCheckPolicyResource(
					"query_data_policy",
					"workspaces/-",
					getQueryDataPolicy(false, 5000000, 1000, 120),
					v1pb.PolicyType_DATA_QUERY,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_policy.query_data_policy"),
					resource.TestCheckResourceAttr("bytebase_policy.query_data_policy", "type", v1pb.PolicyType_DATA_QUERY.String()),
					resource.TestCheckResourceAttr("bytebase_policy.query_data_policy", "query_data_policy.#", "1"),
					resource.TestCheckResourceAttr("bytebase_policy.query_data_policy", "query_data_policy.0.disable_export", "false"),
					resource.TestCheckResourceAttr("bytebase_policy.query_data_policy", "query_data_policy.0.maximum_result_size", "5000000"),
					resource.TestCheckResourceAttr("bytebase_policy.query_data_policy", "query_data_policy.0.maximum_result_rows", "1000"),
					resource.TestCheckResourceAttr("bytebase_policy.query_data_policy", "query_data_policy.0.timeout_in_seconds", "120"),
				),
			},
		},
	})
}

func testAccCheckPolicyDestroy(s *terraform.State) error {
	c, ok := testAccProvider.Meta().(api.Client)
	if !ok {
		return errors.Errorf("cannot get the api client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_policy" {
			continue
		}

		if err := c.DeletePolicy(context.Background(), rs.Primary.ID); err != nil {
			return err
		}
	}

	return nil
}
