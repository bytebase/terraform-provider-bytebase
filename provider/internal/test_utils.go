package internal

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

// GetTestStepForDataSourceList returns the test step for data source test.
func GetTestStepForDataSourceList(
	createResource, dependsOn, dataSourceKey, identifier, fieldKey string,
	expectCount int,
) resource.TestStep {
	resourceKey := fmt.Sprintf("data.%s.%s", dataSourceKey, identifier)
	outputKey := fmt.Sprintf("%s_%s_output_count", dataSourceKey, identifier)

	return resource.TestStep{
		Config: getDataSourceListConfig(createResource, dependsOn, dataSourceKey, identifier, fieldKey, outputKey),
		Check: resource.ComposeTestCheckFunc(
			TestCheckResourceExists(resourceKey),
			resource.TestCheckOutput(outputKey, fmt.Sprintf("%d", expectCount)),
		),
	}
}

func getDataSourceListConfig(resource, dependsOn, dataSourceKey, identifier, fieldKey, outputKey string) string {
	return fmt.Sprintf(`
	%s

	data "%s" "%s" {
		depends_on = [
    		%s
  		]
	}

	output "%s" {
		value = length(data.%s.%s.%s)
	}
	`, resource, dataSourceKey, identifier, dependsOn, outputKey, dataSourceKey, identifier, fieldKey)
}

// GetTestStepForDataSourceListWithParent returns the test step for data source test with an explicit parent.
func GetTestStepForDataSourceListWithParent(
	createResource, dependsOn, dataSourceKey, identifier, fieldKey, parent string,
	expectCount int,
) resource.TestStep {
	resourceKey := fmt.Sprintf("data.%s.%s", dataSourceKey, identifier)
	outputKey := fmt.Sprintf("%s_%s_output_count", dataSourceKey, identifier)

	return resource.TestStep{
		Config: getDataSourceListConfigWithParent(createResource, dependsOn, dataSourceKey, identifier, fieldKey, parent, outputKey),
		Check: resource.ComposeTestCheckFunc(
			TestCheckResourceExists(resourceKey),
			resource.TestCheckOutput(outputKey, fmt.Sprintf("%d", expectCount)),
		),
	}
}

func getDataSourceListConfigWithParent(resource, dependsOn, dataSourceKey, identifier, fieldKey, parent, outputKey string) string {
	return fmt.Sprintf(`
	%s

	data "%s" "%s" {
		parent = "%s"
		depends_on = [
    		%s
  		]
	}

	output "%s" {
		value = length(data.%s.%s.%s)
	}
	`, resource, dataSourceKey, identifier, parent, dependsOn, outputKey, dataSourceKey, identifier, fieldKey)
}

// TestCheckResourceExists will check if the resource exists in the state.
func TestCheckResourceExists(resourceKey string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceKey]

		if !ok {
			return errors.Errorf("Not found: %s", resourceKey)
		}

		if rs.Primary.ID == "" {
			return errors.Errorf("No resource set")
		}

		return nil
	}
}
