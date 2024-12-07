package provider

import (
	"context"
	"fmt"
	"testing"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
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
					"masking_policy",
					"instances/test-sample-instance/databases/employee",
					getMaskingPolicy("salary", "amount", v1pb.MaskingLevel_FULL),
					v1pb.PolicyType_MASKING,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_policy.masking_policy"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_policy", "type", v1pb.PolicyType_MASKING.String()),
					resource.TestCheckResourceAttr("bytebase_policy.masking_policy", "masking_policy.#", "1"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_policy", "masking_policy.0.mask_data.#", "1"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_policy", "masking_policy.0.mask_data.0.table", "salary"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_policy", "masking_policy.0.mask_data.0.column", "amount"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_policy", "masking_policy.0.mask_data.0.masking_level", v1pb.MaskingLevel_FULL.String()),
				),
			},
			{
				Config: testAccCheckPolicyResource(
					"masking_exception_policy",
					"projects/project-sample",
					getMaskingExceptionPolicy("instances/test-sample-instance/databases/employee", "salary", "amount", v1pb.MaskingLevel_PARTIAL),
					v1pb.PolicyType_MASKING_EXCEPTION,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_policy.masking_exception_policy"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exception_policy", "type", v1pb.PolicyType_MASKING_EXCEPTION.String()),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exception_policy", "masking_exception_policy.#", "1"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exception_policy", "masking_exception_policy.0.exceptions.#", "1"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exception_policy", "masking_exception_policy.0.exceptions.0.table", "salary"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exception_policy", "masking_exception_policy.0.exceptions.0.column", "amount"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exception_policy", "masking_exception_policy.0.exceptions.0.masking_level", v1pb.MaskingLevel_PARTIAL.String()),
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

func getMaskingPolicy(table, column string, level v1pb.MaskingLevel) string {
	return fmt.Sprintf(`
	masking_policy {
		mask_data {
			table         = "%s"
			column        = "%s"
			masking_level = "%s"
		}
	}
	`, table, column, level.String())
}

func getMaskingExceptionPolicy(database, table, column string, level v1pb.MaskingLevel) string {
	return fmt.Sprintf(`
	masking_exception_policy {
		exceptions {
			database      = "%s"
			table         = "%s"
			column        = "%s"
			masking_level = "%s"
			member        = "user:ed@bytebase.com"
			action        = "QUERY"
		}
	}
	`, database, table, column, level.String())
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
