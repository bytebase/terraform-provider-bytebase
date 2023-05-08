package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccProjectListDataSource(t *testing.T) {
	identifier := "new_project"
	resourceID := "dev-project"
	title := "dev project"
	key := "BYT"

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
				testAccCheckProjectResource(identifier, resourceID, title, key, api.ProjectWorkflowUI, api.ProjectSchemaVersionSemantic, api.ProjectSchemaChangeDDL),
				fmt.Sprintf("bytebase_project.%s", identifier),
				"bytebase_project_list",
				"after",
				"projects",
				1,
			),
		},
	})
}
