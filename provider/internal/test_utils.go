package internal

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

// GetTestStepForDataSourceList returns the test step for data source test.
func GetTestStepForDataSourceList(
	dataSourceKey, identifier, fieldKey string,
	expectCount int,
) resource.TestStep {
	resourceKey := fmt.Sprintf("data.%s.%s", dataSourceKey, identifier)
	outputKey := fmt.Sprintf("%s_%s_output_count", dataSourceKey, identifier)

	return resource.TestStep{
		Config: getDataSourceListConfig(dataSourceKey, identifier, fieldKey, outputKey),
		Check: resource.ComposeTestCheckFunc(
			TestCheckResourceExists(resourceKey),
			resource.TestCheckOutput(outputKey, fmt.Sprintf("%d", expectCount)),
		),
	}
}

func getDataSourceListConfig(dataSourceKey, identifier, fieldKey, outputKey string) string {
	return fmt.Sprintf(`
	data "%s" "%s" {}

	output "%s" {
		value = length(data.%s.%s.%s)
	}
	`, dataSourceKey, identifier, outputKey, dataSourceKey, identifier, fieldKey)
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
