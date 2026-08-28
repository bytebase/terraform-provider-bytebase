package provider

import (
	"context"
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

func TestResourceProjectWebhookURLPreservedLikePassword(t *testing.T) {
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
	if !urlSchema.Optional || !urlSchema.Computed {
		t.Fatal("webhooks.url should be Optional+Computed like other write-only passwords")
	}
	if urlSchema.StateFunc != nil {
		t.Fatal("webhooks.url must not transform plaintext state")
	}
	if urlSchema.DiffSuppressFunc == nil {
		t.Fatal("webhooks.url should suppress API-omitted values like other write-only passwords")
	}
}

func TestFlattenWebhookListKeepsResourceWebhookURLPlaintext(t *testing.T) {
	plaintext := "https://hooks.example.com/services/customer-secret"

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
	if got := webhook["url"]; got != plaintext {
		t.Fatalf("flattened resource webhook url = %q, want plaintext %q", got, plaintext)
	}
}

func TestFlattenWebhookListPreservesResourceWebhookURL(t *testing.T) {
	plaintext := "https://hooks.example.com/services/customer-secret"

	prior := []interface{}{
		map[string]interface{}{
			"name":  "projects/project-id/webhooks/webhook-id",
			"title": "release alerts",
			"type":  v1pb.WebhookType_SLACK.String(),
			"url":   plaintext,
		},
	}
	raw := flattenWebhookList([]*v1pb.Webhook{{
		Name:  "projects/project-id/webhooks/webhook-id",
		Title: "release alerts",
		Type:  v1pb.WebhookType_SLACK,
	}}, true, prior)

	webhook := raw[0].(map[string]interface{})
	if got := webhook["url"]; got != plaintext {
		t.Fatalf("flattened resource webhook url = %q, want prior plaintext %q", got, plaintext)
	}
}

func TestFlattenWebhookListKeepsDataSourceWebhookURLEmpty(t *testing.T) {
	raw := flattenWebhookList([]*v1pb.Webhook{{
		Name:              "projects/project-id/webhooks/webhook-id",
		Title:             "release alerts",
		Type:              v1pb.WebhookType_SLACK,
		NotificationTypes: []v1pb.Activity_Type{v1pb.Activity_ISSUE_CREATED},
	}}, false)

	if len(raw) != 1 {
		t.Fatalf("flattenWebhookList returned %d webhooks, want 1", len(raw))
	}
	webhook := raw[0].(map[string]interface{})
	if got := webhook["url"]; got != "" {
		t.Fatalf("flattened data source webhook url = %q, want empty write-only value", got)
	}
}

func TestAccProjectWebhookConvergesWhenAPIHidesURL(t *testing.T) {
	const resourceName = "bytebase_project.webhook_state"
	firstURL := "https://hooks.example.com/services/first-secret"
	secondURL := "https://hooks.example.com/services/second-secret"
	rotatedSecondURL := "https://hooks.example.com/services/rotated-second-secret"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckProjectResourceWithWebhooks(firstURL, secondURL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "webhooks.0.url", firstURL),
					resource.TestCheckResourceAttr(resourceName, "webhooks.1.url", secondURL),
				),
			},
			{
				Config:   testAccCheckProjectResourceWithWebhooks(firstURL, secondURL),
				PlanOnly: true,
			},
			{
				Config: testAccCheckProjectResourceWithWebhooks(firstURL, rotatedSecondURL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "webhooks.0.url", firstURL),
					resource.TestCheckResourceAttr(resourceName, "webhooks.1.url", rotatedSecondURL),
				),
			},
			{
				Config:   testAccCheckProjectResourceWithWebhooks(firstURL, rotatedSecondURL),
				PlanOnly: true,
			},
		},
	})
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

func TestResourceProjectIssueLabelsNotComputed(t *testing.T) {
	issueLabels, ok := resourceProjct().Schema["issue_labels"]
	if !ok {
		t.Fatal("issue_labels schema is missing")
	}
	if !issueLabels.Optional {
		t.Fatal("issue_labels should stay Optional")
	}
	if issueLabels.Computed {
		t.Fatal("issue_labels must not be Computed: zero blocks reach the SDK as an empty collection, which it drops before the diff, so a Computed block collection can never be cleared")
	}
}

func TestAccProjectClearIssueLabels(t *testing.T) {
	identifier := "project_clear_issue_labels"
	resourceName := fmt.Sprintf("bytebase_project.%s", identifier)

	resourceID := "test-project-clear-issue-labels"
	title := "test project clear issue labels"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			// create with two labels
			{
				Config: testAccCheckProjectResourceWithSettings(identifier, resourceID, title),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "issue_labels.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
				),
			},
			// removing every block clears the labels, while the labels map, which
			// stays Optional+Computed, keeps what the server holds
			{
				Config: testAccCheckProjectResource(identifier, resourceID, title),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "issue_labels.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
				),
			},
		},
	})
}

func TestAccProjectClearLabels(t *testing.T) {
	identifier := "project_clear_labels"
	resourceName := fmt.Sprintf("bytebase_project.%s", identifier)

	resourceID := "test-project-clear-labels"
	title := "test project clear labels"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckProjectResourceWithSettings(identifier, resourceID, title),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
				),
			},
			// an explicit empty map clears the labels map
			{
				Config: testAccCheckProjectResourceWithEmptyLabels(identifier, resourceID, title),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "0"),
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

func testAccCheckProjectResourceWithWebhooks(firstURL, secondURL string) string {
	return fmt.Sprintf(`
	resource "bytebase_project" "webhook_state" {
		resource_id = "webhook-state"
		title       = "Webhook state"

		webhooks {
			title              = "Issue alerts"
			type               = "SLACK"
			url                = %q
			notification_types = ["ISSUE_CREATED"]
		}

		webhooks {
			title              = "Pipeline alerts"
			type               = "SLACK"
			url                = %q
			notification_types = ["PIPELINE_COMPLETED"]
		}
	}
	`, firstURL, secondURL)
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
			color {
				red   = 1
				green = 0
				blue  = 0
			}
			group = "type"
		}
		issue_labels {
			value = "feature"
			color {
				red   = 0
				green = 1
				blue  = 0
			}
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
			color {
				red   = 1
				green = 0.647059
				blue  = 0
			}
			group = "priority"
		}

		labels = {
			owner = "dba-team"
		}
	}
	`, identifier, resourceID, title)
}

func testAccCheckProjectResourceWithEmptyLabels(identifier, resourceID, title string) string {
	return fmt.Sprintf(`
	resource "bytebase_project" "%s" {
		resource_id    = "%s"
		title          = "%s"

		labels = {}
	}
	`, identifier, resourceID, title)
}
