package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccEnvironmentListDataSource(t *testing.T) {
	identifier := "new-environment"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			internal.GetTestStepForDataSourceList(
				"",
				"",
				"bytebase_environment_list",
				"before",
				"environments",
				0,
			),
			internal.GetTestStepForDataSourceList(
				testAccCheckEnvironmentResource(identifier, "test", 1),
				fmt.Sprintf("bytebase_environment.%s", identifier),
				"bytebase_environment_list",
				"after",
				"environments",
				1,
			),
		},
	})
}
