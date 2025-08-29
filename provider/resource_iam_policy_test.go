package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccIAMPolicy(t *testing.T) {
	identifier := "test_iam"
	projectID := "test-iam-project"
	
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil, // IAM Policy doesn't support delete
		Steps: []resource.TestStep{
			// Create project IAM policy
			{
				Config: testAccCheckProjectIAMPolicyResource(identifier, projectID),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(fmt.Sprintf("bytebase_iam_policy.%s", identifier)),
					resource.TestCheckResourceAttr(fmt.Sprintf("bytebase_iam_policy.%s", identifier), "parent", fmt.Sprintf("projects/%s", projectID)),
					resource.TestCheckResourceAttr(fmt.Sprintf("bytebase_iam_policy.%s", identifier), "iam_policy.#", "1"),
					resource.TestCheckResourceAttr(fmt.Sprintf("bytebase_iam_policy.%s", identifier), "iam_policy.0.binding.#", "1"),
				),
			},
			// Update project IAM policy with multiple bindings
			{
				Config: testAccCheckProjectIAMPolicyResourceWithMultipleBindings(identifier, projectID),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(fmt.Sprintf("bytebase_iam_policy.%s", identifier)),
					resource.TestCheckResourceAttr(fmt.Sprintf("bytebase_iam_policy.%s", identifier), "parent", fmt.Sprintf("projects/%s", projectID)),
					resource.TestCheckResourceAttr(fmt.Sprintf("bytebase_iam_policy.%s", identifier), "iam_policy.0.binding.#", "2"),
				),
			},
		},
	})
}

func TestAccIAMPolicy_Workspace(t *testing.T) {
	identifier := "test_workspace_iam"
	
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil, // IAM Policy doesn't support delete
		Steps: []resource.TestStep{
			// Create workspace IAM policy
			{
				Config: testAccCheckWorkspaceIAMPolicyResource(identifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(fmt.Sprintf("bytebase_iam_policy.%s", identifier)),
					resource.TestCheckResourceAttr(fmt.Sprintf("bytebase_iam_policy.%s", identifier), "parent", "workspaces/-"),
					resource.TestCheckResourceAttr(fmt.Sprintf("bytebase_iam_policy.%s", identifier), "iam_policy.#", "1"),
					resource.TestCheckResourceAttr(fmt.Sprintf("bytebase_iam_policy.%s", identifier), "iam_policy.0.binding.#", "1"),
				),
			},
		},
	})
}

func TestAccIAMPolicy_InvalidInput(t *testing.T) {
	identifier := "invalid_iam"
	
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			// Invalid parent format
			{
				Config: fmt.Sprintf(`
resource "bytebase_iam_policy" "%s" {
	parent = "invalid-parent"
	iam_policy {
		binding {
			role = "roles/test-role"
			members = ["users/test@example.com"]
		}
	}
}
`, identifier),
				ExpectError: regexp.MustCompile("(expected value of parent to match regular expression|Resource id not match|doesn't must any patterns)"),
			},
			// Invalid role format
			{
				Config: fmt.Sprintf(`
resource "bytebase_iam_policy" "%s" {
	parent = "projects/test-project"
	iam_policy {
		binding {
			role = "invalid-role"
			members = ["user:test@example.com"]
		}
	}
}
`, identifier),
				ExpectError: regexp.MustCompile("(invalid role format|role must in roles|doesn't must any patterns.*roles)"),
			},
			// No members
			{
				Config: fmt.Sprintf(`
resource "bytebase_iam_policy" "%s" {
	parent = "projects/test-project"
	iam_policy {
		binding {
			role = "roles/test-role"
			members = []
		}
	}
}
`, identifier),
				ExpectError: regexp.MustCompile("(expected members to have at least|require at least one member|empty members)"),
			},
		},
	})
}

func testAccCheckProjectIAMPolicyResource(identifier, projectID string) string {
	return fmt.Sprintf(`
# Create a project first
resource "bytebase_project" "test_project_%s" {
	resource_id = "%s"
	title       = "Test IAM Project"
}

# Create a role to use in the IAM policy
resource "bytebase_role" "test_role_%s" {
	resource_id = "test-iam-role"
	title       = "Test IAM Role"
	description = "Role for IAM testing"
	permissions = ["bb.permission.database.query"]
}

# Create a user to grant permissions to
resource "bytebase_user" "test_user_%s" {
	email    = "iam-test@example.com"
	title    = "Test IAM User"
	password = "test_password"
	type     = "USER"
}

resource "bytebase_iam_policy" "%s" {
	parent = bytebase_project.test_project_%s.name
	
	iam_policy {
		binding {
			role    = bytebase_role.test_role_%s.name
			members = ["user:iam-test@example.com"]
		}
	}
}
`, identifier, projectID, identifier, identifier, identifier, identifier, identifier)
}

func testAccCheckProjectIAMPolicyResourceWithMultipleBindings(identifier, projectID string) string {
	return fmt.Sprintf(`
# Create a project first
resource "bytebase_project" "test_project_%s" {
	resource_id = "%s"
	title       = "Test IAM Project"
}

# Create roles to use in the IAM policy
resource "bytebase_role" "test_role_%s" {
	resource_id = "test-iam-role"
	title       = "Test IAM Role"
	description = "Role for IAM testing"
	permissions = ["bb.permission.database.query"]
}

resource "bytebase_role" "test_role2_%s" {
	resource_id = "test-iam-role2"
	title       = "Test IAM Role 2"
	description = "Second role for IAM testing"
	permissions = ["bb.permission.database.export"]
}

# Create users to grant permissions to
resource "bytebase_user" "test_user_%s" {
	email    = "iam-test@example.com"
	title    = "Test IAM User"
	password = "test_password"
	type     = "USER"
}

resource "bytebase_user" "test_user2_%s" {
	email    = "iam-test2@example.com"
	title    = "Test IAM User 2"
	password = "test_password"
	type     = "USER"
}

resource "bytebase_iam_policy" "%s" {
	parent = bytebase_project.test_project_%s.name
	
	iam_policy {
		binding {
			role    = bytebase_role.test_role_%s.name
			members = ["user:iam-test@example.com"]
		}
		binding {
			role    = bytebase_role.test_role2_%s.name
			members = [
				"user:iam-test@example.com",
				"user:iam-test2@example.com"
			]
		}
	}
}
`, identifier, projectID, identifier, identifier, identifier, identifier, identifier, identifier, identifier, identifier)
}

func testAccCheckWorkspaceIAMPolicyResource(identifier string) string {
	return fmt.Sprintf(`
# Create a role to use in the IAM policy
resource "bytebase_role" "workspace_role_%s" {
	resource_id = "workspace-iam-role"
	title       = "Workspace IAM Role"
	description = "Role for workspace IAM testing"
	permissions = ["bb.permission.workspace.manage"]
}

# Create a user to grant permissions to
resource "bytebase_user" "workspace_user_%s" {
	email    = "workspace-iam@example.com"
	title    = "Workspace IAM User"
	password = "test_password"
	type     = "USER"
}

resource "bytebase_iam_policy" "%s" {
	parent = "workspaces/-"
	
	iam_policy {
		binding {
			role    = bytebase_role.workspace_role_%s.name
			members = ["user:workspace-iam@example.com"]
		}
	}
}
`, identifier, identifier, identifier, identifier)
}