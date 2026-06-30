package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestReviewRuleHashIncludesPayload(t *testing.T) {
	base := map[string]interface{}{
		"type":   v1pb.SQLReviewRule_COLUMN_REQUIRED.String(),
		"engine": v1pb.Engine_POSTGRES.String(),
		"level":  v1pb.SQLReviewRule_WARNING.String(),
	}

	stringArrayRule := cloneReviewRuleForTest(base)
	stringArrayRule["string_array_payload"] = []interface{}{"id"}
	changedStringArrayRule := cloneReviewRuleForTest(base)
	changedStringArrayRule["string_array_payload"] = []interface{}{"id", "created_ts"}
	if reviewRuleHash(stringArrayRule) == reviewRuleHash(changedStringArrayRule) {
		t.Fatal("review rule hash should change when string_array_payload changes")
	}

	namingRule := map[string]interface{}{
		"type":   v1pb.SQLReviewRule_NAMING_TABLE.String(),
		"engine": v1pb.Engine_POSTGRES.String(),
		"level":  v1pb.SQLReviewRule_WARNING.String(),
		"naming_payload": []interface{}{
			map[string]interface{}{
				"format":     "^[a-z]+$",
				"max_length": 64,
			},
		},
	}
	changedNamingRule := cloneReviewRuleForTest(namingRule)
	changedNamingRule["naming_payload"] = []interface{}{
		map[string]interface{}{
			"format":     "^[a-z][a-z0-9_]+$",
			"max_length": 64,
		},
	}
	if reviewRuleHash(namingRule) == reviewRuleHash(changedNamingRule) {
		t.Fatal("review rule hash should change when naming_payload changes")
	}
}

func TestSetReviewConfigPreservesConfiguredResourcesWhenResponseOmitsResources(t *testing.T) {
	resourceSchema := resourceReviewConfig().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
		"resource_id": "review-config-for-env-test",
		"title":       "Review config for env test",
		"enabled":     true,
		"resources": []interface{}{
			"environments/test",
		},
		"rules": []interface{}{
			map[string]interface{}{
				"type":   v1pb.SQLReviewRule_STATEMENT_WHERE_REQUIRE_SELECT.String(),
				"engine": v1pb.Engine_POSTGRES.String(),
				"level":  v1pb.SQLReviewRule_WARNING.String(),
			},
		},
	})

	diags := setReviewConfig(d, &v1pb.ReviewConfig{
		Name:    "reviewConfigs/review-config-for-env-test",
		Title:   "Review config for env test",
		Enabled: true,
		Rules: []*v1pb.SQLReviewRule{
			{
				Type:   v1pb.SQLReviewRule_STATEMENT_WHERE_REQUIRE_SELECT,
				Engine: v1pb.Engine_POSTGRES,
				Level:  v1pb.SQLReviewRule_WARNING,
			},
		},
	})
	if diags.HasError() {
		t.Fatalf("set review config returned diagnostics: %v", diags)
	}

	resources := d.Get("resources").(*schema.Set)
	if !resources.Contains("environments/test") {
		t.Fatalf("expected configured resources to remain in state, got %#v", resources.List())
	}
}

func cloneReviewRuleForTest(rule map[string]interface{}) map[string]interface{} {
	clone := make(map[string]interface{}, len(rule))
	for key, value := range rule {
		clone[key] = value
	}
	return clone
}

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
		type   = "TABLE_NO_FK"
		engine = "POSTGRES"
		level  = "WARNING"
	}
}
`, identifier),
				ExpectError: regexp.MustCompile(`(expected "resource_id" to not be an empty string|invalid value for resource_id)`),
			},
			// Empty title
			{
				Config: fmt.Sprintf(`
resource "bytebase_review_config" "%s" {
	resource_id = "test-review"
	title       = ""
	enabled     = true
	rules {
		type   = "TABLE_NO_FK"
		engine = "POSTGRES"
		level  = "WARNING"
	}
}
`, identifier),
				ExpectError: regexp.MustCompile(`expected "title" to not be an empty string`),
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
				ExpectError: regexp.MustCompile(`(expected rules to have at least|At least 1 "rules" blocks are required|Missing required argument)`),
			},
			// Invalid engine
			{
				Config: fmt.Sprintf(`
resource "bytebase_review_config" "%s" {
	resource_id = "test-review"
	title       = "Test Review"
	enabled     = true
	rules {
		type   = "TABLE_NO_FK"
		engine = "INVALID_ENGINE"
		level  = "WARNING"
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
		type   = "TABLE_NO_FK"
		engine = "POSTGRES"
		level  = "INVALID_LEVEL"
	}
}
`, identifier),
				ExpectError: regexp.MustCompile(`(expected rules.0.level to be one of|must be one of)`),
			},
			// Invalid attached resource
			{
				Config: fmt.Sprintf(`
resource "bytebase_review_config" "%s" {
	resource_id = "test-review"
	title       = "Test Review"
	enabled     = true
	resources = [
		"instances/test",
	]
	rules {
		type   = "STATEMENT_WHERE_REQUIRE_SELECT"
		engine = "POSTGRES"
		level  = "WARNING"
	}
}
`, identifier),
				ExpectError: regexp.MustCompile(`invalid resource, only support projects/\{id\} or environments/\{id\}`),
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
		type   = "NAMING_TABLE"
		engine = "POSTGRES"
		level  = "%s"
		naming_payload {
			max_length = 64
		}
	}

	rules {
		type   = "NAMING_COLUMN"
		engine = "MYSQL"
		level  = "%s"
		naming_payload {
			max_length = 64
		}
	}
}
`, identifier, resourceID, title, enabled,
		v1pb.SQLReviewRule_WARNING.String(),
		v1pb.SQLReviewRule_ERROR.String())
}

func testAccCheckReviewConfigResourceWithMoreRules(identifier, resourceID, title string) string {
	return fmt.Sprintf(`
resource "bytebase_review_config" "%s" {
	resource_id = "%s"
	title       = "%s"
	enabled     = true

	rules {
		type   = "NAMING_TABLE"
		engine = "POSTGRES"
		level  = "%s"
		naming_payload {
			max_length = 64
		}
	}

	rules {
		type   = "NAMING_COLUMN"
		engine = "MYSQL"
		level  = "%s"
		naming_payload {
			max_length = 64
		}
	}

	rules {
		type   = "STATEMENT_WHERE_REQUIRE_SELECT"
		engine = "POSTGRES"
		level  = "%s"
	}
}
`, identifier, resourceID, title,
		v1pb.SQLReviewRule_WARNING.String(),
		v1pb.SQLReviewRule_ERROR.String(),
		v1pb.SQLReviewRule_WARNING.String())
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
