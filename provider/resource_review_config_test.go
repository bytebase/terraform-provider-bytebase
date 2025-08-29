package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccReviewConfig(t *testing.T) {
	identifier := "test_review"
	resourceName := fmt.Sprintf("bytebase_review_config.%s", identifier)

	resourceID := "test-review-config"
	title := "Test Review Config"
	titleUpdated := "Updated Test Review Config"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckReviewConfigDestroy,
		Steps: []resource.TestStep{
			// Create review config with basic rules
			{
				Config: testAccCheckReviewConfigResource(identifier, resourceID, title, true),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_id", resourceID),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "2"),
				),
			},
			// Update review config
			{
				Config: testAccCheckReviewConfigResource(identifier, resourceID, titleUpdated, false),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_id", resourceID),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "2"),
				),
			},
			// Update with more rules
			{
				Config: testAccCheckReviewConfigResourceWithMoreRules(identifier, resourceID, titleUpdated),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_id", resourceID),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "3"),
				),
			},
		},
	})
}

func TestAccReviewConfig_InvalidInput(t *testing.T) {
	identifier := "invalid_review"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckReviewConfigDestroy,
		Steps: []resource.TestStep{
			// Empty resource_id
			{
				Config: fmt.Sprintf(`
resource "bytebase_review_config" "%s" {
	resource_id = ""
	title       = "Test Review"
	enabled     = true
	rules {
		type    = "naming.table"
		engine  = "POSTGRES"
		level   = "WARNING"
		payload = "{}"
	}
}
`, identifier),
				ExpectError: regexp.MustCompile("(expected \"resource_id\" to not be an empty string|invalid value for resource_id)"),
			},
			// Empty title
			{
				Config: fmt.Sprintf(`
resource "bytebase_review_config" "%s" {
	resource_id = "test-review"
	title       = ""
	enabled     = true
	rules {
		type    = "naming.table"
		engine  = "POSTGRES"
		level   = "WARNING"
		payload = "{}"
	}
}
`, identifier),
				ExpectError: regexp.MustCompile("expected \"title\" to not be an empty string"),
			},
			// No rules
			{
				Config: fmt.Sprintf(`
resource "bytebase_review_config" "%s" {
	resource_id = "test-review"
	title       = "Test Review"
	enabled     = true
}
`, identifier),
				ExpectError: regexp.MustCompile("(expected rules to have at least|At least 1 \"rules\" blocks are required|Missing required argument)"),
			},
			// Invalid engine
			{
				Config: fmt.Sprintf(`
resource "bytebase_review_config" "%s" {
	resource_id = "test-review"
	title       = "Test Review"
	enabled     = true
	rules {
		type    = "naming.table"
		engine  = "INVALID_ENGINE"
		level   = "WARNING"
		payload = "{}"
	}
}
`, identifier),
				ExpectError: regexp.MustCompile(`(expected rules.0.engine to be one of|invalid value for engine)`),
			},
			// Invalid level
			{
				Config: fmt.Sprintf(`
resource "bytebase_review_config" "%s" {
	resource_id = "test-review"
	title       = "Test Review"
	enabled     = true
	rules {
		type    = "naming.table"
		engine  = "POSTGRES"
		level   = "INVALID_LEVEL"
		payload = "{}"
	}
}
`, identifier),
				ExpectError: regexp.MustCompile(`(expected rules.0.level to be one of|must be one of)`),
			},
		},
	})
}

func testAccCheckReviewConfigResource(identifier, resourceID, title string, enabled bool) string {
	return fmt.Sprintf(`
resource "bytebase_review_config" "%s" {
	resource_id = "%s"
	title       = "%s"
	enabled     = %t
	
	rules {
		type    = "naming.table"
		engine  = "POSTGRES"
		level   = "%s"
		payload = "{\"format\":{\"maxLength\":64}}"
		comment = "Table naming rule"
	}
	
	rules {
		type    = "naming.column"
		engine  = "MYSQL"
		level   = "%s"
		payload = "{\"format\":{\"maxLength\":64}}"
		comment = "Column naming rule"
	}
}
`, identifier, resourceID, title, enabled,
		v1pb.SQLReviewRuleLevel_WARNING.String(),
		v1pb.SQLReviewRuleLevel_ERROR.String())
}

func testAccCheckReviewConfigResourceWithMoreRules(identifier, resourceID, title string) string {
	return fmt.Sprintf(`
resource "bytebase_review_config" "%s" {
	resource_id = "%s"
	title       = "%s"
	enabled     = true
	
	rules {
		type    = "naming.table"
		engine  = "POSTGRES"
		level   = "%s"
		payload = "{\"format\":{\"maxLength\":64}}"
		comment = "Table naming rule"
	}
	
	rules {
		type    = "naming.column"
		engine  = "MYSQL"
		level   = "%s"
		payload = "{\"format\":{\"maxLength\":64}}"
		comment = "Column naming rule"
	}
	
	rules {
		type    = "statement.select"
		engine  = "POSTGRES"
		level   = "%s"
		payload = "{}"
		comment = "Select statement rule"
	}
}
`, identifier, resourceID, title,
		v1pb.SQLReviewRuleLevel_WARNING.String(),
		v1pb.SQLReviewRuleLevel_ERROR.String(),
		v1pb.SQLReviewRuleLevel_WARNING.String())
}

func testAccCheckReviewConfigDestroy(s *terraform.State) error {
	c, ok := testAccProvider.Meta().(api.Client)
	if !ok {
		return errors.Errorf("cannot get the api client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_review_config" {
			continue
		}

		if err := c.DeleteReviewConfig(context.Background(), rs.Primary.ID); err != nil {
			return err
		}
	}

	return nil
}
