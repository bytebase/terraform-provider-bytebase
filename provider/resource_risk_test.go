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

func TestAccRisk(t *testing.T) {
	identifier := "test_risk"
	resourceName := fmt.Sprintf("bytebase_risk.%s", identifier)

	title := "Test Risk"
	titleUpdated := "Updated Test Risk"
	source := v1pb.Risk_DDL.String()
	level := 300
	levelUpdated := 200

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRiskDestroy,
		Steps: []resource.TestStep{
			// Create risk
			{
				Config: testAccCheckRiskResource(identifier, title, source, level, true),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "source", source),
					resource.TestCheckResourceAttr(resourceName, "level", fmt.Sprintf("%d", level)),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
				),
			},
			// Update risk
			{
				Config: testAccCheckRiskResource(identifier, titleUpdated, source, levelUpdated, false),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "source", source),
					resource.TestCheckResourceAttr(resourceName, "level", fmt.Sprintf("%d", levelUpdated)),
					resource.TestCheckResourceAttr(resourceName, "active", "false"),
				),
			},
			// Update with different source
			{
				Config: testAccCheckRiskResource(identifier, titleUpdated, v1pb.Risk_DML.String(), level, true),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "source", v1pb.Risk_DML.String()),
					resource.TestCheckResourceAttr(resourceName, "level", fmt.Sprintf("%d", level)),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
				),
			},
		},
	})
}

func TestAccRisk_AllSources(t *testing.T) {
	sources := []string{
		v1pb.Risk_DDL.String(),
		v1pb.Risk_DML.String(),
		v1pb.Risk_CREATE_DATABASE.String(),
		v1pb.Risk_REQUEST_ROLE.String(),
		v1pb.Risk_DATA_EXPORT.String(),
	}

	for _, source := range sources {
		t.Run(source, func(t *testing.T) {
			identifier := fmt.Sprintf("test_risk_%s", source)
			resourceName := fmt.Sprintf("bytebase_risk.%s", identifier)

			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(t)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccCheckRiskDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccCheckRiskResource(identifier, fmt.Sprintf("Test Risk %s", source), source, 300, true),
						Check: resource.ComposeTestCheckFunc(
							internal.TestCheckResourceExists(resourceName),
							resource.TestCheckResourceAttr(resourceName, "source", source),
						),
					},
				},
			})
		})
	}
}

func TestAccRisk_InvalidInput(t *testing.T) {
	identifier := "invalid_risk"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRiskDestroy,
		Steps: []resource.TestStep{
			// Empty title
			{
				Config: fmt.Sprintf(`
resource "bytebase_risk" "%s" {
	title     = ""
	source    = "%s"
	level     = 300
	active    = true
	condition = "{\"expressions\":[{\"title\":\"High risk database\",\"expression\":\"resource.database == 'production'\"}]}"
}
`, identifier, v1pb.Risk_DDL.String()),
				ExpectError: regexp.MustCompile(`expected "title" to not be an empty string`),
			},
			// Invalid source
			{
				Config: fmt.Sprintf(`
resource "bytebase_risk" "%s" {
	title     = "Test Risk"
	source    = "INVALID_SOURCE"
	level     = 300
	active    = true
	condition = "{\"expressions\":[{\"title\":\"High risk database\",\"expression\":\"resource.database == 'production'\"}]}"
}
`, identifier),
				ExpectError: regexp.MustCompile(`(expected source to be one of|must be one of)`),
			},
			// Invalid level
			{
				Config: fmt.Sprintf(`
resource "bytebase_risk" "%s" {
	title     = "Test Risk"
	source    = "%s"
	level     = 150
	active    = true
	condition = "{\"expressions\":[{\"title\":\"High risk database\",\"expression\":\"resource.database == 'production'\"}]}"
}
`, identifier, v1pb.Risk_DDL.String()),
				ExpectError: regexp.MustCompile(`(expected level to be one of|must be one of)`),
			},
		},
	})
}

func testAccCheckRiskResource(identifier, title, source string, level int, active bool) string {
	// Different conditions based on source type
	condition := ""
	switch source {
	case v1pb.Risk_DDL.String():
		condition = `{
			"expressions": [{
				"title": "DDL Risk",
				"expression": "resource.database_name == \"production\""
			}]
		}`
	case v1pb.Risk_DML.String():
		condition = `{
			"expressions": [{
				"title": "DML Risk",
				"expression": "level == \"HIGH\" && source == \"UI\""
			}]
		}`
	case v1pb.Risk_CREATE_DATABASE.String():
		condition = `{
			"expressions": [{
				"title": "Database Creation Risk",
				"expression": "environment_id == \"prod\""
			}]
		}`
	case v1pb.Risk_REQUEST_ROLE.String():
		condition = `{
			"expressions": [{
				"title": "Role Request Risk",
				"expression": "role contains \"OWNER\""
			}]
		}`
	case v1pb.Risk_DATA_EXPORT.String():
		condition = `{
			"expressions": [{
				"title": "Data Export Risk",
				"expression": "database contains \"customer\""
			}]
		}`
	default:
		condition = `{
			"expressions": [{
				"title": "Default Risk",
				"expression": "true"
			}]
		}`
	}

	return fmt.Sprintf(`
resource "bytebase_risk" "%s" {
	title     = "%s"
	source    = "%s"
	level     = %d
	active    = %t
	condition = jsonencode(%s)
}
`, identifier, title, source, level, active, condition)
}

func testAccCheckRiskDestroy(s *terraform.State) error {
	c, ok := testAccProvider.Meta().(api.Client)
	if !ok {
		return errors.Errorf("cannot get the api client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_risk" {
			continue
		}

		if err := c.DeleteRisk(context.Background(), rs.Primary.ID); err != nil {
			return err
		}
	}

	return nil
}
