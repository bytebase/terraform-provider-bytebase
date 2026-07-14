package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccInstanceListDataSource(t *testing.T) {
	identifier := "staging_instance"

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
			{
				Config: fmt.Sprintf(`
%s

data "bytebase_instance_list" "after" {
	depends_on = [
		bytebase_instance.%s
	]
}
`, testAccCheckInstanceResourceWithLabels(identifier, "staging-instance", "staging instance", "POSTGRES", "environments/staging", "staging", "platform"), identifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists("data.bytebase_instance_list.after"),
					resource.TestCheckResourceAttr("data.bytebase_instance_list.after", "instances.#", "1"),
					resource.TestCheckResourceAttr("data.bytebase_instance_list.after", "instances.0.labels.%", "2"),
					resource.TestCheckResourceAttr("data.bytebase_instance_list.after", "instances.0.labels.environment", "staging"),
					resource.TestCheckResourceAttr("data.bytebase_instance_list.after", "instances.0.labels.team", "platform"),
				),
			},
		},
	})
}
