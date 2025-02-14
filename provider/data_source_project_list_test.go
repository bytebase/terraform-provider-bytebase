package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccProjectListDataSource(t *testing.T) {
	identifier := "new_project"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			internal.GetTestStepForDataSourceList(
				"",
				"",
				"bytebase_project_list",
				"before",
				"projects",
				0,
			),
			internal.GetTestStepForDataSourceList(
				testAccCheckProjectResource(identifier, "dev-project", "dev project"),
				fmt.Sprintf("bytebase_project.%s", identifier),
				"bytebase_project_list",
				"after",
				"projects",
				1,
			),
		},
	})
}
