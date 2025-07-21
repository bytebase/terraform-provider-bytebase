package provider

import (
	"context"
	"fmt"
	"testing"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
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
					"masking_exception_policy",
					"projects/project-sample",
					getMaskingExceptionPolicy("instances/test-sample-instance/databases/employee", "salary", "amount"),
					v1pb.PolicyType_MASKING_EXCEPTION,
				),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("bytebase_policy.masking_exception_policy"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exception_policy", "type", v1pb.PolicyType_MASKING_EXCEPTION.String()),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exception_policy", "masking_exception_policy.#", "1"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exception_policy", "masking_exception_policy.0.exceptions.#", "1"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exception_policy", "masking_exception_policy.0.exceptions.0.table", "salary"),
					resource.TestCheckResourceAttr("bytebase_policy.masking_exception_policy", "masking_exception_policy.0.exceptions.0.column", "amount"),
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

func getMaskingExceptionPolicy(database, table, column string) string {
	return fmt.Sprintf(`
	masking_exception_policy {
		exceptions {
			database      = "%s"
			table         = "%s"
			column        = "%s"
			member        = "user:ed@bytebase.com"
			action        = "QUERY"
		}
	}
	`, database, table, column)
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
