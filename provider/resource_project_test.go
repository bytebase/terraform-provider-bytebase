package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestResourceProjectWebhookURLStoredAsHash(t *testing.T) {
	webhooks, ok := resourceProjct().Schema["webhooks"]
	if !ok {
		t.Fatal("webhooks schema is missing")
	}
	elem, ok := webhooks.Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("webhooks Elem = %T, want *schema.Resource", webhooks.Elem)
	}
	urlSchema, ok := elem.Schema["url"]
	if !ok {
		t.Fatal("webhooks.url schema is missing")
	}
	if urlSchema.WriteOnly {
		t.Fatal("webhooks.url must not use WriteOnly; nested WriteOnly causes non-converging inline webhook diffs")
	}
	if !urlSchema.Sensitive {
		t.Fatal("webhooks.url should be Sensitive so plaintext is hidden in CLI output")
	}
	if urlSchema.StateFunc == nil {
		t.Fatal("webhooks.url should hash plaintext before storing it in Terraform state")
	}

	plaintext := "https://hooks.example.com/services/customer-secret"
	sum := sha256.Sum256([]byte(plaintext))
	want := hex.EncodeToString(sum[:])
	if got := urlSchema.StateFunc(plaintext); got != want {
		t.Fatalf("url StateFunc = %q, want sha256 %q", got, want)
	}
}

func TestFlattenWebhookListHashesResourceWebhookURL(t *testing.T) {
	plaintext := "https://hooks.example.com/services/customer-secret"
	sum := sha256.Sum256([]byte(plaintext))
	want := hex.EncodeToString(sum[:])

	raw := flattenWebhookList([]*v1pb.Webhook{{
		Name:              "projects/project-id/webhooks/webhook-id",
		Title:             "release alerts",
		Type:              v1pb.WebhookType_SLACK,
		Url:               plaintext,
		NotificationTypes: []v1pb.Activity_Type{v1pb.Activity_ISSUE_CREATED},
	}}, true)

	if len(raw) != 1 {
		t.Fatalf("flattenWebhookList returned %d webhooks, want 1", len(raw))
	}
	webhook := raw[0].(map[string]interface{})
	if got := webhook["url"]; got != want {
		t.Fatalf("flattened resource webhook url = %q, want sha256 %q", got, want)
	}
	if got := webhook["url"]; got == plaintext {
		t.Fatal("flattened resource webhook url stores plaintext")
	}
}

func TestFlattenWebhookListKeepsDataSourceWebhookURLPlaintext(t *testing.T) {
	plaintext := "https://hooks.example.com/services/customer-secret"

	raw := flattenWebhookList([]*v1pb.Webhook{{
		Name:              "projects/project-id/webhooks/webhook-id",
		Title:             "release alerts",
		Type:              v1pb.WebhookType_SLACK,
		Url:               plaintext,
		NotificationTypes: []v1pb.Activity_Type{v1pb.Activity_ISSUE_CREATED},
	}}, false)

	if len(raw) != 1 {
		t.Fatalf("flattenWebhookList returned %d webhooks, want 1", len(raw))
	}
	webhook := raw[0].(map[string]interface{})
	if got := webhook["url"]; got != plaintext {
		t.Fatalf("flattened data source webhook url = %q, want plaintext %q", got, plaintext)
	}
}

func TestAccProject(t *testing.T) {
	identifier := "new_project"
	resourceName := fmt.Sprintf("bytebase_project.%s", identifier)

	resourceID := "test-project"
	title := "test project"
	titleUpdated := fmt.Sprintf("%s-updated", title)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			// resource create
			{
				Config: testAccCheckProjectResource(identifier, resourceID, title),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", title),
				),
			},
			// resource updated
			{
				Config: testAccCheckProjectResource(identifier, resourceID, titleUpdated),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
				),
			},
		},
	})
}

func TestAccProjectWithSettings(t *testing.T) {
	identifier := "project_with_settings"
	resourceName := fmt.Sprintf("bytebase_project.%s", identifier)

	resourceID := "test-project-settings"
	title := "test project with settings"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			// resource create with settings
			{
				Config: testAccCheckProjectResourceWithSettings(identifier, resourceID, title),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "enforce_sql_review", "true"),
					resource.TestCheckResourceAttr(resourceName, "require_issue_approval", "true"),
					resource.TestCheckResourceAttr(resourceName, "require_plan_check_no_error", "false"),
					resource.TestCheckResourceAttr(resourceName, "allow_request_role", "true"),
					resource.TestCheckResourceAttr(resourceName, "force_issue_labels", "true"),
					resource.TestCheckResourceAttr(resourceName, "issue_labels.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "labels.environment", "test"),
					resource.TestCheckResourceAttr(resourceName, "labels.team", "platform"),
				),
			},
			// resource update settings
			{
				Config: testAccCheckProjectResourceWithSettingsUpdated(identifier, resourceID, title),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enforce_sql_review", "false"),
					resource.TestCheckResourceAttr(resourceName, "require_issue_approval", "false"),
					resource.TestCheckResourceAttr(resourceName, "issue_labels.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "1"),
				),
			},
		},
	})
}

func testAccCheckProjectDestroy(s *terraform.State) error {
	c, ok := testAccProvider.Meta().(api.Client)
	if !ok {
		return errors.Errorf("cannot get the api client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_project" {
			continue
		}

		if err := c.DeleteProject(context.Background(), rs.Primary.ID); err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckProjectResource(identifier, resourceID, title string) string {
	return fmt.Sprintf(`
	resource "bytebase_project" "%s" {
		resource_id    = "%s"
		title          = "%s"
	}
	`, identifier, resourceID, title)
}

func testAccCheckProjectResourceWithSettings(identifier, resourceID, title string) string {
	return fmt.Sprintf(`
	resource "bytebase_project" "%s" {
		resource_id    = "%s"
		title          = "%s"

		enforce_sql_review          = true
		require_issue_approval      = true
		require_plan_check_no_error = false
		allow_request_role          = true
		force_issue_labels          = true

		issue_labels {
			value = "bug"
			color = "#FF0000"
			group = "type"
		}
		issue_labels {
			value = "feature"
			color = "#00FF00"
			group = "type"
		}

		labels = {
			environment = "test"
			team        = "platform"
		}
	}
	`, identifier, resourceID, title)
}

func testAccCheckProjectResourceWithSettingsUpdated(identifier, resourceID, title string) string {
	return fmt.Sprintf(`
	resource "bytebase_project" "%s" {
		resource_id    = "%s"
		title          = "%s"

		enforce_sql_review          = false
		require_issue_approval      = false
		require_plan_check_no_error = true
		allow_request_role          = false
		force_issue_labels          = false

		issue_labels {
			value = "urgent"
			color = "#FFA500"
			group = "priority"
		}

		labels = {
			owner = "dba-team"
		}
	}
	`, identifier, resourceID, title)
}
