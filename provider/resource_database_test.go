package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccDatabase(t *testing.T) {
	identifier := "test_db"
	resourceName := fmt.Sprintf("bytebase_database.%s", identifier)

	instanceID := "test-instance"
	databaseName := "test-database"
	projectName := "projects/test-project"
	environmentName := "environments/test"

	databaseFullName := fmt.Sprintf("instances/%s/databases/%s", instanceID, databaseName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			// resource create
			{
				Config: testAccCheckDatabaseResource(identifier, databaseFullName, projectName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", databaseFullName),
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "environment", environmentName),
				),
			},
			// resource update - just check it still works
			{
				Config: testAccCheckDatabaseResource(identifier, databaseFullName, projectName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", databaseFullName),
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "environment", environmentName),
				),
			},
		},
	})
}

func TestAccDatabase_WithLabels(t *testing.T) {
	identifier := "test_db_labels"
	resourceName := fmt.Sprintf("bytebase_database.%s", identifier)

	instanceID := "test-instance"
	databaseName := "test-database-labels"
	projectName := "projects/test-project"
	environmentName := "environments/test"

	databaseFullName := fmt.Sprintf("instances/%s/databases/%s", instanceID, databaseName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			// resource create with labels
			{
				Config: testAccCheckDatabaseResourceWithLabels(identifier, databaseFullName, projectName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", databaseFullName),
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "environment", environmentName),
					resource.TestCheckResourceAttr(resourceName, "labels.env", "test"),
					resource.TestCheckResourceAttr(resourceName, "labels.app", "terraform"),
				),
			},
			// resource update labels
			{
				Config: testAccCheckDatabaseResourceWithLabelsUpdated(identifier, databaseFullName, projectName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", databaseFullName),
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "environment", environmentName),
					resource.TestCheckResourceAttr(resourceName, "labels.env", "production"),
					resource.TestCheckResourceAttr(resourceName, "labels.app", "terraform-updated"),
					resource.TestCheckResourceAttr(resourceName, "labels.tier", "backend"),
				),
			},
		},
	})
}

func TestAccDatabase_InvalidInput(t *testing.T) {
	identifier := "invalid_db"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			// Invalid database name format
			{
				Config:      testAccCheckDatabaseInvalidName(identifier),
				ExpectError: regexp.MustCompile(`(expected value of name to match regular expression|Resource id not match|doesn't must any patterns)`),
			},
			// Invalid project name format
			{
				Config:      testAccCheckDatabaseInvalidProject(identifier),
				ExpectError: regexp.MustCompile(`(expected value of project to match regular expression|Resource id not match|doesn't must any patterns)`),
			},
			// Invalid environment name format
			{
				Config:      testAccCheckDatabaseInvalidEnvironment(identifier),
				ExpectError: regexp.MustCompile(`(expected value of environment to match regular expression|Resource id not match|doesn't must any patterns)`),
			},
		},
	})
}

func testAccCheckDatabaseResource(identifier, name, project, environment string) string {
	// Extract instance ID from database name
	instanceID := ""
	if strings.HasPrefix(name, "instances/") {
		parts := strings.Split(name, "/")
		if len(parts) >= 2 {
			instanceID = parts[1]
		}
	}

	// Extract project ID from project name
	projectID := ""
	if strings.HasPrefix(project, "projects/") {
		projectID = strings.TrimPrefix(project, "projects/")
	}

	// Extract environment ID from environment name
	environmentID := ""
	if strings.HasPrefix(environment, "environments/") {
		environmentID = strings.TrimPrefix(environment, "environments/")
	}

	return fmt.Sprintf(`
# Create environment first
resource "bytebase_environment" "test_env_%s" {
	resource_id = "%s"
	title       = "Test Environment"
	order       = 0
}

# Create project
resource "bytebase_project" "test_project_%s" {
	resource_id = "%s"
	title       = "Test Project"
}

# Create instance with a default database
resource "bytebase_instance" "test_instance_%s" {
	resource_id = "%s"
	title       = "Test Instance"
	engine      = "POSTGRES"
	environment = bytebase_environment.test_env_%s.name

	data_sources {
		id       = "admin"
		type     = "ADMIN"
		username = "postgres"
		host     = "127.0.0.1"
		port     = "5432"
	}
}

# Now manage the database
resource "bytebase_database" "%s" {
	name        = "%s"
	project     = bytebase_project.test_project_%s.name
	environment = bytebase_environment.test_env_%s.name

	depends_on = [bytebase_instance.test_instance_%s]
}
`, identifier, environmentID, identifier, projectID, identifier, instanceID, identifier, identifier, name, identifier, identifier, identifier)
}

func testAccCheckDatabaseResourceWithLabels(identifier, name, project, environment string) string {
	// Extract instance ID from database name
	instanceID := ""
	if strings.HasPrefix(name, "instances/") {
		parts := strings.Split(name, "/")
		if len(parts) >= 2 {
			instanceID = parts[1]
		}
	}

	// Extract project ID from project name
	projectID := ""
	if strings.HasPrefix(project, "projects/") {
		projectID = strings.TrimPrefix(project, "projects/")
	}

	// Extract environment ID from environment name
	environmentID := ""
	if strings.HasPrefix(environment, "environments/") {
		environmentID = strings.TrimPrefix(environment, "environments/")
	}

	return fmt.Sprintf(`
# Create environment first
resource "bytebase_environment" "test_env_%s" {
	resource_id = "%s"
	title       = "Test Environment"
	order       = 0
}

# Create project
resource "bytebase_project" "test_project_%s" {
	resource_id = "%s"
	title       = "Test Project"
}

# Create instance with a default database
resource "bytebase_instance" "test_instance_%s" {
	resource_id = "%s"
	title       = "Test Instance"
	engine      = "POSTGRES"
	environment = bytebase_environment.test_env_%s.name

	data_sources {
		id       = "admin"
		type     = "ADMIN"
		username = "postgres"
		host     = "127.0.0.1"
		port     = "5432"
	}
}

# Now manage the database with labels
resource "bytebase_database" "%s" {
	name        = "%s"
	project     = bytebase_project.test_project_%s.name
	environment = bytebase_environment.test_env_%s.name
	labels = {
		env = "test"
		app = "terraform"
	}

	depends_on = [bytebase_instance.test_instance_%s]
}
`, identifier, environmentID, identifier, projectID, identifier, instanceID, identifier, identifier, name, identifier, identifier, identifier)
}

func testAccCheckDatabaseResourceWithLabelsUpdated(identifier, name, project, environment string) string {
	// Extract instance ID from database name
	instanceID := ""
	if strings.HasPrefix(name, "instances/") {
		parts := strings.Split(name, "/")
		if len(parts) >= 2 {
			instanceID = parts[1]
		}
	}

	// Extract project ID from project name
	projectID := ""
	if strings.HasPrefix(project, "projects/") {
		projectID = strings.TrimPrefix(project, "projects/")
	}

	// Extract environment ID from environment name
	environmentID := ""
	if strings.HasPrefix(environment, "environments/") {
		environmentID = strings.TrimPrefix(environment, "environments/")
	}

	return fmt.Sprintf(`
# Create environment first
resource "bytebase_environment" "test_env_%s" {
	resource_id = "%s"
	title       = "Test Environment"
	order       = 0
}

# Create project
resource "bytebase_project" "test_project_%s" {
	resource_id = "%s"
	title       = "Test Project"
}

# Create instance with a default database
resource "bytebase_instance" "test_instance_%s" {
	resource_id = "%s"
	title       = "Test Instance"
	engine      = "POSTGRES"
	environment = bytebase_environment.test_env_%s.name

	data_sources {
		id       = "admin"
		type     = "ADMIN"
		username = "postgres"
		host     = "127.0.0.1"
		port     = "5432"
	}
}

# Now manage the database with updated labels
resource "bytebase_database" "%s" {
	name        = "%s"
	project     = bytebase_project.test_project_%s.name
	environment = bytebase_environment.test_env_%s.name
	labels = {
		env  = "production"
		app  = "terraform-updated"
		tier = "backend"
	}

	depends_on = [bytebase_instance.test_instance_%s]
}
`, identifier, environmentID, identifier, projectID, identifier, instanceID, identifier, identifier, name, identifier, identifier, identifier)
}

func testAccCheckDatabaseInvalidName(identifier string) string {
	return fmt.Sprintf(`
resource "bytebase_database" "%s" {
	name        = "invalid-database-name"
	project     = "projects/test"
	environment = "environments/test"
}
`, identifier)
}

func testAccCheckDatabaseInvalidProject(identifier string) string {
	return fmt.Sprintf(`
resource "bytebase_database" "%s" {
	name        = "instances/test/databases/db"
	project     = "invalid-project"
	environment = "environments/test"
}
`, identifier)
}

func testAccCheckDatabaseInvalidEnvironment(identifier string) string {
	return fmt.Sprintf(`
resource "bytebase_database" "%s" {
	name        = "instances/test/databases/db"
	project     = "projects/test"
	environment = "invalid-env"
}
`, identifier)
}

func testAccCheckDatabaseDestroy(_ *terraform.State) error {
	// In the mock implementation, databases are not actually deleted
	// They remain as part of the instance. This is fine for testing
	// as we're primarily testing the Terraform resource lifecycle.
	// In a real environment, the database deletion would be handled differently.
	return nil
}
