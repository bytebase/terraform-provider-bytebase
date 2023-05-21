package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccDatabaseListDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			internal.GetTestStepForDataSourceList(
				"",
				"",
				"bytebase_database_list",
				"before",
				"databases",
				0,
			),
			internal.GetTestStepForDataSourceList(
				testAccCheckInstanceResource("dev_instance_with_db", "dev-instance-with-db", "Dev Instance", "POSTGRES", "dev-env"),
				"bytebase_instance.dev_instance_with_db",
				"bytebase_database_list",
				"after",
				"databases",
				1,
			),
		},
	})
}
