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
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.conditions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.flow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.flow.0.steps.#", "2"),
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
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.conditions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.flow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_flow.0.rules.0.flow.0.steps.#", "1"),
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
				),
			},
			// Update workspace profile setting with domains
			{
				Config: testAccCheckWorkspaceProfileSettingWithDomains(identifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.domains.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "workspace_profile.0.enforce_identity_domain", "true"),
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
					resource.TestCheckResourceAttr(resourceName, "classification.0.classification_from_config", "true"),
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

func TestAccSetting_PasswordRestriction(t *testing.T) {
	identifier := "test_password"
	resourceName := fmt.Sprintf("bytebase_setting.%s", identifier)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			// Create password restriction setting
			{
				Config: testAccCheckPasswordRestrictionSetting(identifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "settings/PASSWORD_RESTRICTION"),
					resource.TestCheckResourceAttr(resourceName, "password_restriction.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "password_restriction.0.min_length", "10"),
					resource.TestCheckResourceAttr(resourceName, "password_restriction.0.require_number", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_restriction.0.require_letter", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_restriction.0.require_uppercase_letter", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_restriction.0.require_special_character", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_restriction.0.password_rotation_in_seconds", "7776000"),
				),
			},
		},
	})
}

func TestAccSetting_SQLQueryRestriction(t *testing.T) {
	identifier := "test_sql_query"
	resourceName := fmt.Sprintf("bytebase_setting.%s", identifier)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			// Create SQL query restriction setting
			{
				Config: testAccCheckSQLQueryRestrictionSetting(identifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "settings/SQL_RESULT_SIZE_LIMIT"),
					resource.TestCheckResourceAttr(resourceName, "sql_query_restriction.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sql_query_restriction.0.maximum_result_size", "1048576"),
					resource.TestCheckResourceAttr(resourceName, "sql_query_restriction.0.maximum_result_rows", "1000"),
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
			conditions {
				source = "DDL"
				level  = "LOW"
			}
			flow {
				title       = "Test"
				description = "Test"
				steps {
					role = "roles/test"
				}
			}
		}
	}
}
`, identifier),
				ExpectError: regexp.MustCompile("(expected value of name to match regular expression|Resource id not match|doesn't must any patterns)"),
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
				ExpectError: regexp.MustCompile("(Missing required argument|Blocks of type \"conditions\" are required|missing expected|Insufficient flow blocks|At least 1 \"flow\" blocks are required)"),
			},
			// Invalid role format in approval flow
			{
				Config: fmt.Sprintf(`
resource "bytebase_setting" "%s" {
	name = "settings/WORKSPACE_APPROVAL"
	approval_flow {
		rules {
			conditions {
				source = "DDL"
				level  = "LOW"
			}
			flow {
				title       = "Test"
				description = "Test"
				steps {
					role = "invalid-role"
				}
			}
		}
	}
}
`, identifier),
				ExpectError: regexp.MustCompile("(invalid role name|Resource id not match|doesn't must any patterns.*roles)"),
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
				ExpectError: regexp.MustCompile("invalid environment id"),
			},
			// Missing classification ID
			{
				Config: fmt.Sprintf(`
resource "bytebase_setting" "%s" {
	name = "settings/DATA_CLASSIFICATION"
	classification {
		title                      = "Test Classification"
		classification_from_config = true
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
				ExpectError: regexp.MustCompile("(id is required for classification|Missing required argument)"),
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
			conditions {
				source = "DDL"
				level  = "LOW"
			}
			flow {
				title       = "DDL Approval Flow"
				description = "Approval flow for DDL operations"
				steps {
					role = bytebase_role.approval_role_%s.name
				}
				steps {
					role = bytebase_role.approval_role_%s.name
				}
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
			conditions {
				source = "DDL"
				level  = "HIGH"
			}
			conditions {
				source = "DML"
				level  = "MODERATE"
			}
			flow {
				title       = "Updated Approval Flow"
				description = "Updated approval flow"
				steps {
					role = bytebase_role.approval_role_%s.name
				}
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
		classification_from_config = true
		
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

func testAccCheckPasswordRestrictionSetting(identifier string) string {
	return fmt.Sprintf(`
resource "bytebase_setting" "%s" {
	name = "settings/PASSWORD_RESTRICTION"
	password_restriction {
		min_length                         = 10
		require_number                     = true
		require_letter                     = true
		require_uppercase_letter           = true
		require_special_character          = true
		require_reset_password_for_first_login = true
		password_rotation_in_seconds       = 7776000
	}
}
`, identifier)
}

func testAccCheckSQLQueryRestrictionSetting(identifier string) string {
	return fmt.Sprintf(`
resource "bytebase_setting" "%s" {
	name = "settings/SQL_RESULT_SIZE_LIMIT"
	sql_query_restriction {
		maximum_result_size = 1048576
		maximum_result_rows = 1000
	}
}
`, identifier)
}