package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccInstanceListDataSource(t *testing.T) {
	identifier := "new_instance"
	name := "dev instance"
	engine := "POSTGRES"
	host := "127.0.0.1"
	environment := "dev"

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
				"bytebase_instance_list",
				"before",
				"instances",
				0,
			),
			internal.GetTestStepForDataSourceList(
				testAccCheckInstanceResource(identifier, name, engine, host, environment),
				fmt.Sprintf("bytebase_instance.%s", identifier),
				"bytebase_instance_list",
				"after",
				"instances",
				1,
			),
		},
	})
}
