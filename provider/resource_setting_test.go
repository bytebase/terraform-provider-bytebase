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
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.source", "CHANGE_DATABASE"),
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
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.source", "CHANGE_DATABASE"),
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
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.maximum_request_expiration_in_seconds", "86400"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.maximum_role_expiration_in_seconds", "86400"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.enable_audit_log_stdout", "true"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.announcement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.announcement.0.text", "Test announcement"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.announcement.0.theme.0.background.0.red", "1"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.announcement.0.theme.0.background.0.green", "0.968627"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.announcement.0.theme.0.background.0.blue", "0.878431"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.announcement.0.theme.0.text.0.red", "0.592157"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.announcement.0.theme.0.text.0.green", "0.352941"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.announcement.0.theme.0.text.0.blue", "0.086275"),
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

func TestSettingEnvironmentColorBlockConfig(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "bytebase_setting" "environments" {
	name = "settings/ENVIRONMENT"

	environment_setting {
		environment {
			id        = "prod"
			title     = "Prod"
			protected = true
			color {
				red   = 1
				green = 0
				blue  = 0
			}
		}
	}
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
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
			source    = "CHANGE_DATABASE"
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
			source    = "CHANGE_DATABASE"
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
			color {
				red   = 1
				green = 0
				blue  = 0
			}
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
			source    = "CHANGE_DATABASE"
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
			source    = "CHANGE_DATABASE"
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
		enforce_identity_domain               = true
		maximum_request_expiration_in_seconds = 86400
		maximum_role_expiration_in_seconds    = 86400
		enable_audit_log_stdout               = true
		announcement {
			text = "Test announcement"
			link = "https://example.com/announcement"
			theme {
				background {
					red   = 1
					green = 0.968627
					blue  = 0.878431
				}
				text {
					red   = 0.592157
					green = 0.352941
					blue  = 0.086275
				}
			}
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
			title = "Public"
			level = 0
		}
		levels {
			title = "Sensitive"
			level = 1
		}
		levels {
			title = "Confidential"
			level = 2
		}

		classifications {
			id    = "0-0"
			title = "Email Address"
			level = 1
		}
		classifications {
			id    = "0-1"
			title = "SSN"
			level = 2
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
			color {
				red   = 0
				green = 1
				blue  = 0
			}
			protected = false
		}
		environment {
			id        = "staging"
			title     = "Staging"
			color {
				red   = 1
				green = 1
				blue  = 0
			}
			protected = true
		}
		environment {
			id        = "prod"
			title     = "Production"
			color {
				red   = 1
				green = 0
				blue  = 0
			}
			protected = true
		}
	}
}
`, identifier)
}
