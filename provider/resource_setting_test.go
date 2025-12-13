package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccSetting_WorkspaceApproval(t *testing.T) {
	identifier := "test_workspace_approval"
	resourceName := fmt.Sprintf("bytebase_setting.%s", identifier)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil, // Settings don't support delete
		Steps: []resource.TestStep{
			// Create workspace approval setting
			{
				Config: testAccCheckWorkspaceApprovalSetting(identifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "settings/WORKSPACE_APPROVAL"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.source", "DDL"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.condition", "request.risk <= 100"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.flow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.flow.0.roles.#", "2"),
				),
			},
			// Update workspace approval setting
			{
				Config: testAccCheckWorkspaceApprovalSettingUpdated(identifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "settings/WORKSPACE_APPROVAL"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.source", "DDL"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.condition", "request.risk > 100"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.flow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.flow.0.roles.#", "1"),
				),
			},
		},
	})
}

func TestAccSetting_WorkspaceProfile(t *testing.T) {
	identifier := "test_workspace_profile"
	resourceName := fmt.Sprintf("bytebase_setting.%s", identifier)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil, // Settings don't support delete
		Steps: []resource.TestStep{
			// Create workspace profile setting
			{
				Config: testAccCheckWorkspaceProfileSetting(identifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "settings/WORKSPACE_PROFILE"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.external_url", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.disallow_signup", "true"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.disallow_password_signin", "false"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.database_change_mode", "PIPELINE"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.enable_audit_log_stdout", "false"),
				),
			},
			// Update workspace profile setting with domains
			{
				Config: testAccCheckWorkspaceProfileSettingWithDomains(identifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.domains.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.enforce_identity_domain", "true"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.enable_audit_log_stdout", "true"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.announcement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.announcement.0.text", "Test announcement"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.announcement.0.level", "INFO"),
				),
			},
		},
	})
}

func TestAccSetting_DataClassification(t *testing.T) {
	identifier := "test_classification"
	resourceName := fmt.Sprintf("bytebase_setting.%s", identifier)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			// Create data classification setting
			{
				Config: testAccCheckDataClassificationSetting(identifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "settings/DATA_CLASSIFICATION"),
					resource.TestCheckResourceAttr(resourceName, "classification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "classification.0.id", "test-classification"),
					resource.TestCheckResourceAttr(resourceName, "classification.0.title", "Test Classification"),
				),
			},
		},
	})
}

func TestAccSetting_SemanticTypes(t *testing.T) {
	identifier := "test_semantic"
	resourceName := fmt.Sprintf("bytebase_setting.%s", identifier)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			// Create semantic types setting
			{
				Config: testAccCheckSemanticTypesSetting(identifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "settings/SEMANTIC_TYPES"),
				),
			},
			// Update with different mask algorithms
			{
				Config: testAccCheckSemanticTypesSettingWithAlgorithms(identifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "settings/SEMANTIC_TYPES"),
				),
			},
		},
	})
}

func TestAccSetting_Environment(t *testing.T) {
	identifier := "test_env_setting"
	resourceName := fmt.Sprintf("bytebase_setting.%s", identifier)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			// Create environment setting
			{
				Config: testAccCheckEnvironmentSetting(identifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "settings/ENVIRONMENT"),
					resource.TestCheckResourceAttr(resourceName, "environment_setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "environment_setting.0.environment.#", "3"),
				),
			},
		},
	})
}

func TestAccSetting_InvalidInput(t *testing.T) {
	identifier := "invalid_setting"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			// Invalid setting name
			{
				Config: fmt.Sprintf(`
resource "bytebase_setting" "%s" {
	name = "settings/INVALID_SETTING"
	approval_flow {
		rules {
			source    = "DDL"
			condition = "request.risk <= 100"
			flow {
				title       = "Test"
				description = "Test"
				roles = ["roles/test"]
			}
		}
	}
}
`, identifier),
				ExpectError: regexp.MustCompile(`(expected value of name to match regular expression|Resource id not match|doesn't must any patterns)`),
			},
			// Missing required fields for approval flow
			{
				Config: fmt.Sprintf(`
resource "bytebase_setting" "%s" {
	name = "settings/WORKSPACE_APPROVAL"
	approval_flow {
		rules {
		}
	}
}
`, identifier),
				ExpectError: regexp.MustCompile(`(Missing required argument|missing expected|Insufficient flow blocks|At least 1 "flow" blocks are required)`),
			},
			// Invalid role format in approval flow
			{
				Config: fmt.Sprintf(`
resource "bytebase_setting" "%s" {
	name = "settings/WORKSPACE_APPROVAL"
	approval_flow {
		rules {
			source    = "DDL"
			condition = "request.risk <= 100"
			flow {
				title       = "Test"
				description = "Test"
				roles = ["invalid-role"]
			}
		}
	}
}
`, identifier),
				ExpectError: regexp.MustCompile(`(invalid role name|Resource id not match|doesn't must any patterns.*roles)`),
			},
			// Invalid environment ID
			{
				Config: fmt.Sprintf(`
resource "bytebase_setting" "%s" {
	name = "settings/ENVIRONMENT"
	environment_setting {
		environment {
			id        = "invalid environment id"
			title     = "Test"
			color     = "#FF0000"
			protected = true
		}
	}
}
`, identifier),
				ExpectError: regexp.MustCompile(`invalid environment id`),
			},
			// Missing classification ID
			{
				Config: fmt.Sprintf(`
resource "bytebase_setting" "%s" {
	name = "settings/DATA_CLASSIFICATION"
	classification {
		title                      = "Test Classification"
		levels {
			id          = "level1"
			title       = "Level 1"
			description = "Test"
		}
		classifications {
			title       = "Class 1"
			description = "Test"
		}
	}
}
`, identifier),
				ExpectError: regexp.MustCompile(`(id is required for classification|Missing required argument)`),
			},
		},
	})
}

func testAccCheckWorkspaceApprovalSetting(identifier string) string {
	return fmt.Sprintf(`
# Create prerequisite role
resource "bytebase_role" "approval_role_%s" {
	resource_id = "approval-test-role"
	title       = "Approval Test Role"
	description = "Role for approval testing"
	permissions = ["bb.permission.database.query"]
}

resource "bytebase_setting" "%s" {
	name = "settings/WORKSPACE_APPROVAL"
	approval_flow {
		rules {
			source    = "DDL"
			condition = "request.risk <= 100"
			flow {
				title       = "DDL Approval Flow"
				description = "Approval flow for DDL operations"
				roles = [
					bytebase_role.approval_role_%s.name,
					bytebase_role.approval_role_%s.name
				]
			}
		}
	}
}
`, identifier, identifier, identifier, identifier)
}

func testAccCheckWorkspaceApprovalSettingUpdated(identifier string) string {
	return fmt.Sprintf(`
# Create prerequisite role
resource "bytebase_role" "approval_role_%s" {
	resource_id = "approval-test-role"
	title       = "Approval Test Role"
	description = "Role for approval testing"
	permissions = ["bb.permission.database.query"]
}

resource "bytebase_setting" "%s" {
	name = "settings/WORKSPACE_APPROVAL"
	approval_flow {
		rules {
			source    = "DDL"
			condition = "request.risk > 100"
			flow {
				title       = "Updated Approval Flow"
				description = "Updated approval flow"
				roles = [bytebase_role.approval_role_%s.name]
			}
		}
	}
}
`, identifier, identifier, identifier)
}

func testAccCheckWorkspaceProfileSetting(identifier string) string {
	return fmt.Sprintf(`
resource "bytebase_setting" "%s" {
	name = "settings/WORKSPACE_PROFILE"
	workspace_profile {
		external_url             = "https://example.com"
		disallow_signup          = true
		disallow_password_signin = false
		database_change_mode     = "PIPELINE"
		token_duration_in_seconds = 3600
		enable_audit_log_stdout  = false
	}
}
`, identifier)
}

func testAccCheckWorkspaceProfileSettingWithDomains(identifier string) string {
	return fmt.Sprintf(`
resource "bytebase_setting" "%s" {
	name = "settings/WORKSPACE_PROFILE"
	workspace_profile {
		external_url                       = "https://example.com"
		disallow_signup                    = true
		disallow_password_signin           = false
		database_change_mode               = "EDITOR"
		domains                            = ["example.com", "test.com"]
		enforce_identity_domain            = true
		token_duration_in_seconds          = 7200
		maximum_role_expiration_in_seconds = 86400
		enable_audit_log_stdout            = true
		announcement {
			text  = "Test announcement"
			link  = "https://example.com/announcement"
			level = "INFO"
		}
	}
}
`, identifier)
}

func testAccCheckDataClassificationSetting(identifier string) string {
	return fmt.Sprintf(`
resource "bytebase_setting" "%s" {
	name = "settings/DATA_CLASSIFICATION"
	classification {
		id                         = "test-classification"
		title                      = "Test Classification"
		
		levels {
			id          = "public"
			title       = "Public"
			description = "Public data"
		}
		levels {
			id          = "sensitive"
			title       = "Sensitive"
			description = "Sensitive data"
		}
		levels {
			id          = "confidential"
			title       = "Confidential"
			description = "Confidential data"
		}
		
		classifications {
			id          = "email"
			title       = "Email Address"
			description = "Email addresses"
			level       = "sensitive"
		}
		classifications {
			id          = "ssn"
			title       = "SSN"
			description = "Social Security Numbers"
			level       = "confidential"
		}
	}
}
`, identifier)
}

func testAccCheckSemanticTypesSetting(identifier string) string {
	return fmt.Sprintf(`
resource "bytebase_setting" "%s" {
	name = "settings/SEMANTIC_TYPES"
	semantic_types {
		id          = "email"
		title       = "Email"
		description = "Email address"
		algorithm {
			full_mask {
				substitution = "******"
			}
		}
	}
	semantic_types {
		id          = "phone"
		title       = "Phone"
		description = "Phone number"
	}
}
`, identifier)
}

func testAccCheckSemanticTypesSettingWithAlgorithms(identifier string) string {
	return fmt.Sprintf(`
resource "bytebase_setting" "%s" {
	name = "settings/SEMANTIC_TYPES"
	semantic_types {
		id          = "ssn"
		title       = "SSN"
		description = "Social Security Number"
		algorithm {
			range_mask {
				slices {
					start        = 0
					end          = 5
					substitution = "*"
				}
			}
		}
	}
	semantic_types {
		id          = "credit_card"
		title       = "Credit Card"
		description = "Credit Card Number"
		algorithm {
			inner_outer_mask {
				prefix_len   = 4
				suffix_len   = 4
				substitution = "*"
				type         = "INNER"
			}
		}
	}
	semantic_types {
		id          = "password"
		title       = "Password"
		description = "User Password"
		algorithm {
			md5_mask {
				salt = "test-salt"
			}
		}
	}
}
`, identifier)
}

func testAccCheckEnvironmentSetting(identifier string) string {
	return fmt.Sprintf(`
resource "bytebase_setting" "%s" {
	name = "settings/ENVIRONMENT"
	environment_setting {
		environment {
			id        = "dev"
			title     = "Development"
			color     = "#00FF00"
			protected = false
		}
		environment {
			id        = "staging"
			title     = "Staging"
			color     = "#FFFF00"
			protected = true
		}
		environment {
			id        = "prod"
			title     = "Production"
			color     = "#FF0000"
			protected = true
		}
	}
}
`, identifier)
}
